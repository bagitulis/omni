import * as os from "os";

interface ProcessMemory {
  rss_mb: number;
  vms_mb: number;
  percent: number;
}

interface ProcessCPU {
  percent: number;
}

interface SystemMemory {
  total_gb: number;
  available_gb: number;
  used_gb: number;
  percent: number;
}

interface SystemCPU {
  percent: number;
  count: number;
}

interface ResourceData {
  backend: {
    memory: ProcessMemory;
    cpu: ProcessCPU;
    threads: number;
    pid: number;
    uptime: number;
  };
  system: {
    memory: SystemMemory;
    cpu: SystemCPU;
  };
  timestamp: string;
}

let lastCPUUsage: NodeJS.CpuUsage | null = null;
let lastCheck = Date.now();

/**
 * Calculate CPU percentage for the current process
 */
function calculateProcessCPUPercent(): number {
  const currentCPUUsage = process.cpuUsage();
  const currentTime = Date.now();

  if (!lastCPUUsage) {
    lastCPUUsage = currentCPUUsage;
    lastCheck = currentTime;
    return 0;
  }

  const elapsedMs = currentTime - lastCheck; // milliseconds
  const userDiff = currentCPUUsage.user - lastCPUUsage.user; // microseconds
  const systemDiff = currentCPUUsage.system - lastCPUUsage.system; // microseconds
  const totalDiff = (userDiff + systemDiff) / 1000; // Convert to milliseconds

  // CPU percentage = (total cpu time / elapsed time) / num_cpus * 100
  const numCPUs = os.cpus().length;
  const cpuPercent = (totalDiff / elapsedMs / numCPUs) * 100;

  lastCPUUsage = currentCPUUsage;
  lastCheck = currentTime;

  return Math.min(cpuPercent, 100); // Cap at 100%
}

/**
 * Get current system CPU usage percentage (approximation)
 */
function getSystemCPUUsagePercent(): number {
  const cpus = os.cpus();
  let totalIdle = 0;
  let totalTick = 0;

  for (const cpu of cpus) {
    for (const type in cpu.times) {
      totalTick += cpu.times[type as keyof typeof cpu.times];
    }
    totalIdle += cpu.times.idle;
  }

  const idle = totalIdle / cpus.length;
  const total = totalTick / cpus.length;
  const usage = 100 - ~~((100 * idle) / total);

  return Math.max(0, Math.min(usage, 100));
}

/**
 * Get resource usage for backend and system
 */
export function getResourceUsage(): ResourceData {
  const memUsage = process.memoryUsage();
  const systemMem = os.totalmem();
  const freeMem = os.freemem();
  const usedMem = systemMem - freeMem;

  // Calculate percentages
  const processMemPercent = (memUsage.heapUsed / memUsage.heapTotal) * 100;
  const systemMemPercent = (usedMem / systemMem) * 100;

  const processCPUPercent = calculateProcessCPUPercent();
  const systemCPUPercent = getSystemCPUUsagePercent();

  const data: ResourceData = {
    backend: {
      memory: {
        rss_mb: Math.round((memUsage.rss / 1024 / 1024) * 100) / 100,
        vms_mb: Math.round((memUsage.heapTotal / 1024 / 1024) * 100) / 100,
        percent: Math.round(processMemPercent * 100) / 100,
      },
      cpu: {
        percent: Math.round(processCPUPercent * 100) / 100,
      },
      threads: 1, // Node.js doesn't expose thread count easily
      pid: process.pid,
      uptime: Math.round(process.uptime()),
    },
    system: {
      memory: {
        total_gb: Math.round((systemMem / 1024 / 1024 / 1024) * 100) / 100,
        available_gb: Math.round((freeMem / 1024 / 1024 / 1024) * 100) / 100,
        used_gb: Math.round((usedMem / 1024 / 1024 / 1024) * 100) / 100,
        percent: Math.round(systemMemPercent * 100) / 100,
      },
      cpu: {
        percent: Math.round(systemCPUPercent * 100) / 100,
        count: os.cpus().length,
      },
    },
    timestamp: new Date().toISOString(),
  };

  return data;
}
