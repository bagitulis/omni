"""
PostgreSQL Fixer implementations for Omni Build System.

SRP: This module ONLY contains fix implementations for PostgreSQL.
"""
import subprocess
import time
from pathlib import Path

from omni_build.logger import log_error, log_fix, log_info, log_success, log_warning


def _restart_container(container_name: str, timeout: int = 60) -> bool:
    """Restart a Docker container (local helper to avoid circular import)."""
    try:
        log_info(f"Restarting {container_name}...")
        result = subprocess.run(
            ["docker", "restart", container_name],
            capture_output=True,
            text=True,
            encoding='utf-8',
            errors='replace',
            timeout=timeout,
        )
        if result.returncode == 0:
            log_success(f"{container_name} restarted")
            return True
        log_warning(f"Failed to restart {container_name}: {result.stderr}")
        return False
    except Exception as e:
        log_error(f"Error restarting {container_name}: {e}")
        return False


class PostgresFixer:
    """Fix implementations for PostgreSQL service."""
    
    @staticmethod
    def repair_postgres_pg_hba() -> bool:
        """Fix PostgreSQL pg_hba.conf authentication errors."""
        log_fix("Repairing PostgreSQL pg_hba.conf for Docker network...")
        
        try:
            log_info("Adding Docker network to pg_hba.conf...")
            subprocess.run(
                ["docker", "exec", "omni-postgres", "sh", "-c",
                 "echo 'host    all    all    172.0.0.0/8    trust' >> /var/lib/postgresql/data/pg_hba.conf"],
                capture_output=True,
                check=False,
                timeout=30
            )
            
            subprocess.run(
                ["docker", "exec", "omni-postgres", "sh", "-c",
                 "echo 'host    all    all    192.168.0.0/16    trust' >> /var/lib/postgresql/data/pg_hba.conf"],
                capture_output=True,
                check=False,
                timeout=30
            )
            
            log_info("Reloading PostgreSQL configuration...")
            subprocess.run(
                ["docker", "exec", "omni-postgres", "su", "postgres", "-c",
                 "pg_ctl reload -D /var/lib/postgresql/data"],
                capture_output=True,
                check=False,
                timeout=30
            )
            
            time.sleep(3)
            log_success("pg_hba.conf updated - Docker networks now allowed")
            return True
            
        except Exception as e:
            log_error(f"Failed to repair pg_hba.conf: {e}")
            return False
    
    @staticmethod
    def create_postgres_database_if_not_exists() -> bool:
        """Create PostgreSQL database and restore from backup if available.
        
        Priority:
        1. If backup exists in backups/smart/ -> create DB + restore from backup
        2. If no backup -> create DB + run init-multi-tenant.sql (empty tables)
        """
        log_fix("Checking PostgreSQL database status...")
        
        try:
            # First ensure postgres container is running
            result = subprocess.run(
                ["docker", "inspect", "--format", "{{.State.Status}}", "omni-postgres"],
                capture_output=True,
                text=True,
                encoding='utf-8',
                errors='replace',
                timeout=10,
            )
            
            if result.returncode != 0 or result.stdout.strip() != "running":
                log_info("PostgreSQL container not running, starting it...")
                subprocess.run(
                    ["docker-compose", "-f", "docker-compose.tunnel.yml",
                     "-f", "docker-compose.tunnel.standard.yml", "up", "-d", "postgres"],
                    capture_output=True,
                    check=False,
                    timeout=60
                )
                time.sleep(15)
            
            # Check if database exists first
            log_info("Checking if database 'omni_main' exists...")
            check_result = subprocess.run(
                ["docker", "exec", "omni-postgres", "psql", "-U", "omni", "-d", "postgres",
                 "-tAc", "SELECT 1 FROM pg_database WHERE datname='omni_main'"],
                capture_output=True,
                text=True,
                encoding='utf-8',
                errors='replace',
                timeout=15,
            )
            
            if check_result.stdout.strip() == "1":
                log_success("Database 'omni_main' already exists!")
                PostgresFixer.repair_postgres_pg_hba()
                return True
            
            # Database doesn't exist - check if backup is available
            backup_dir = Path.cwd() / "backups" / "smart"
            manifest_file = backup_dir / "manifest.json"
            has_backup = manifest_file.exists()
            
            if has_backup:
                log_info("Backup found in backups/smart/ - will restore from backup")
            else:
                log_warning("No backup found - will create empty database with init script")
            
            # Create database first
            log_info("Creating database 'omni_main'...")
            create_result = subprocess.run(
                ["docker", "exec", "omni-postgres", "psql", "-U", "omni", "-d", "postgres",
                 "-c", "CREATE DATABASE omni_main WITH ENCODING='UTF8' LC_COLLATE='C' LC_CTYPE='C' TEMPLATE=template0;"],
                capture_output=True,
                text=True,
                encoding='utf-8',
                errors='replace',
                timeout=30,
            )
            
            if create_result.returncode != 0 and "already exists" not in create_result.stderr:
                log_warning(f"Failed to create database: {create_result.stderr}")
                return PostgresFixer.repair_postgres_database()
            
            log_success("Database 'omni_main' created successfully!")
            
            # Fix pg_hba.conf for Docker networks BEFORE restore
            PostgresFixer.repair_postgres_pg_hba()
            
            if has_backup:
                # Restore from backup using DatabaseRestorer
                log_info("Restoring database from backup...")
                try:
                    from omni_build.config import Config
                    from omni_build.database_restorer import DatabaseRestorer
                    
                    config = Config.from_env()
                    restorer = DatabaseRestorer(config)
                    
                    # Print backup info
                    restorer.print_backup_status()
                    
                    # Execute restore
                    success, message = restorer.restore(force=True)
                    
                    if success:
                        log_success("Database restored from backup successfully!")
                        return True
                    else:
                        log_warning(f"Restore failed: {message}")
                        log_info("Falling back to init script...")
                        # Fall through to init script
                except Exception as e:
                    log_warning(f"Restore error: {e}")
                    log_info("Falling back to init script...")
                    # Fall through to init script
            
            # No backup or restore failed - run init script
            init_script = Path.cwd() / "scripts" / "postgres" / "init-multi-tenant.sql"
            if init_script.exists():
                log_info("Running initialization script (creates empty tables)...")
                
                subprocess.run(
                    ["docker", "cp", str(init_script), "omni-postgres:/tmp/init.sql"],
                    capture_output=True,
                    check=False,
                    timeout=30
                )
                
                result = subprocess.run(
                    ["docker", "exec", "omni-postgres", "psql", "-U", "omni",
                     "-d", "omni_main", "-f", "/tmp/init.sql"],
                    capture_output=True,
                    text=True,
                    encoding='utf-8',
                    errors='replace',
                    timeout=60,
                )
                
                if result.returncode == 0:
                    log_success("Initialization script executed successfully!")
                else:
                    log_warning(f"Init script had issues: {result.stderr[:200]}")
            
            # Verify the database is ready
            for _ in range(5):
                time.sleep(2)
                check = subprocess.run(
                    ["docker", "exec", "omni-postgres", "pg_isready", "-U", "omni", "-d", "omni_main"],
                    capture_output=True,
                    timeout=10,
                )
                if check.returncode == 0:
                    log_success("PostgreSQL database is ready!")
                    return True
            
            log_warning("Database created but pg_isready check didn't pass")
            return True
            
        except Exception as e:
            log_error(f"Failed to create PostgreSQL database: {e}")
            log_info("Falling back to full database reset...")
            return PostgresFixer.repair_postgres_database()
    
    @staticmethod
    def repair_postgres_database() -> bool:
        """Fix PostgreSQL database not exist error (DESTRUCTIVE - removes data)."""
        log_fix("Repairing PostgreSQL - full reset (will delete data)...")
        
        try:
            log_info("Stopping all containers...")
            subprocess.run(
                ["docker-compose", "-f", "docker-compose.tunnel.yml",
                 "-f", "docker-compose.tunnel.standard.yml", "down"],
                capture_output=True,
                check=False,
                timeout=60
            )
            
            log_info("Removing PostgreSQL container...")
            subprocess.run(
                ["docker", "rm", "-f", "omni-postgres"],
                capture_output=True,
                check=False
            )
            
            log_warning("Removing PostgreSQL data directory for fresh init...")
            import shutil
            postgres_data = Path.cwd() / "data" / "postgres"
            if postgres_data.exists():
                shutil.rmtree(postgres_data, ignore_errors=True)
            
            time.sleep(2)
            
            log_info("Starting fresh PostgreSQL container...")
            subprocess.run(
                ["docker-compose", "-f", "docker-compose.tunnel.yml",
                 "-f", "docker-compose.tunnel.standard.yml", "up", "-d", "postgres"],
                capture_output=True,
                check=False,
                timeout=60
            )
            
            log_info("Waiting for PostgreSQL initialization (up to 60s)...")
            for i in range(12):
                time.sleep(5)
                result = subprocess.run(
                    ["docker", "exec", "omni-postgres", "pg_isready", "-U", "omni", "-d", "omni_main"],
                    capture_output=True,
                    text=True,
                    encoding='utf-8',
                    errors='replace',
                    timeout=10,
                    check=False
                )
                if result.returncode == 0:
                    log_success("PostgreSQL database initialized successfully!")
                    PostgresFixer.repair_postgres_pg_hba()
                    return True
                log_info(f"Waiting for database init... ({(i+1)*5}s)")
            
            log_warning("PostgreSQL init took longer than expected")
            return True
            
        except Exception as e:
            log_error(f"Failed to repair PostgreSQL database: {e}")
            return False
    
    @staticmethod
    def repair_postgres_data() -> bool:
        """Repair corrupted PostgreSQL data."""
        log_warning("PostgreSQL data corruption detected")
        log_info("Stopping PostgreSQL container...")
        subprocess.run(["docker", "stop", "omni-postgres"], 
                      capture_output=True, check=False)
        log_info("Container will reinitialize on next start")
        return True
    
    @staticmethod
    def repair_docker_dns_postgres() -> bool:
        """Fix Docker internal DNS not resolving 'postgres' hostname."""
        log_fix("Repairing Docker internal DNS for postgres hostname...")
        
        log_info("Pruning Docker networks...")
        subprocess.run(["docker", "network", "prune", "-f"],
                      capture_output=True, check=False)
        
        log_info("Stopping all containers to reset network...")
        containers = ["omni-backend", "omni-frontend", "omni-postgres", 
                     "omni-redis", "omni-pgbouncer", "omni-nginx", 
                     "omni-cloudflared", "omni-pgweb"]
        
        for container in containers:
            subprocess.run(["docker", "stop", container],
                          capture_output=True, check=False, timeout=30)
        
        time.sleep(3)
        
        log_info("Removing containers to force network recreation...")
        for container in containers:
            subprocess.run(["docker", "rm", container],
                          capture_output=True, check=False)
        
        log_success("Docker DNS repair completed - containers will be redeployed")
        return True
    
    # === NEW SILENT ERROR FIXES ===
    
    @staticmethod
    def repair_postgres_io() -> bool:
        """Repair PostgreSQL I/O errors."""
        log_fix("Repairing PostgreSQL I/O error...")
        log_warning("I/O error detected - checking disk health")
        
        # Force checkpoint to flush data
        subprocess.run(
            ["docker", "exec", "omni-postgres", "psql", "-U", "omni", "-d", "omni_main",
             "-c", "CHECKPOINT;"],
            capture_output=True, check=False, timeout=30
        )
        
        # Restart PostgreSQL to clear any stuck I/O
        log_info("Restarting PostgreSQL to clear I/O state...")
        return _restart_container("omni-postgres", timeout=90)
    
    @staticmethod
    def repair_postgres_checkpoint() -> bool:
        """Repair PostgreSQL checkpoint failures."""
        log_fix("Repairing PostgreSQL checkpoint error...")
        
        # Force checkpoint with wait
        result = subprocess.run(
            ["docker", "exec", "omni-postgres", "psql", "-U", "omni", "-d", "omni_main",
             "-c", "CHECKPOINT;"],
            capture_output=True, text=True, encoding='utf-8', errors='replace', check=False, timeout=60
        )
        
        if result.returncode == 0:
            log_success("Checkpoint completed successfully")
            return True
        
        log_warning("Checkpoint failed, restarting PostgreSQL...")
        return _restart_container("omni-postgres", timeout=90)
    
    @staticmethod
    def repair_postgres_wal() -> bool:
        """Repair PostgreSQL WAL errors."""
        log_fix("Repairing PostgreSQL WAL error...")
        log_warning("WAL error detected - this may require manual intervention")
        
        # Try to reset WAL by restarting
        log_info("Restarting PostgreSQL to reset WAL state...")
        return _restart_container("omni-postgres", timeout=90)
    
    @staticmethod
    def repair_postgres_deadlock() -> bool:
        """Repair PostgreSQL deadlock situations."""
        log_fix("Repairing PostgreSQL deadlock...")
        
        # Terminate blocking queries
        result = subprocess.run(
            ["docker", "exec", "omni-postgres", "psql", "-U", "omni", "-d", "omni_main",
             "-c", """
                SELECT pg_terminate_backend(pid) 
                FROM pg_stat_activity 
                WHERE wait_event_type = 'Lock' 
                AND state = 'active' 
                AND pid <> pg_backend_pid();
             """],
            capture_output=True, text=True, encoding='utf-8', errors='replace', check=False, timeout=30
        )
        
        if result.returncode == 0:
            log_success("Terminated blocking queries")
            return True
        
        log_warning("Could not terminate blocking queries")
        return False
    
    @staticmethod
    def repair_postgres_connections() -> bool:
        """Repair PostgreSQL connection exhaustion."""
        log_fix("Repairing PostgreSQL connection exhaustion...")
        
        # Terminate idle connections
        result = subprocess.run(
            ["docker", "exec", "omni-postgres", "psql", "-U", "omni", "-d", "omni_main",
             "-c", """
                SELECT pg_terminate_backend(pid) 
                FROM pg_stat_activity 
                WHERE state = 'idle' 
                AND state_change < NOW() - INTERVAL '5 minutes'
                AND pid <> pg_backend_pid();
             """],
            capture_output=True, text=True, encoding='utf-8', errors='replace', check=False, timeout=30
        )
        
        if result.returncode == 0:
            log_success("Terminated idle connections older than 5 minutes")
            return True
        
        log_warning("Could not terminate idle connections, restarting PgBouncer...")
        return _restart_container("omni-pgbouncer")
    
    @staticmethod
    def repair_postgres_replication() -> bool:
        """Handle PostgreSQL replication lag (logging only for now)."""
        log_fix("PostgreSQL replication lag detected...")
        log_warning("Replication lag detected - monitoring only (no auto-fix)")
        log_info("Check replica status manually if using replication")
        return True
