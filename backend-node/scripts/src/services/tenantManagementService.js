"use strict";
var __assign = (this && this.__assign) || function () {
    __assign = Object.assign || function(t) {
        for (var s, i = 1, n = arguments.length; i < n; i++) {
            s = arguments[i];
            for (var p in s) if (Object.prototype.hasOwnProperty.call(s, p))
                t[p] = s[p];
        }
        return t;
    };
    return __assign.apply(this, arguments);
};
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
var __generator = (this && this.__generator) || function (thisArg, body) {
    var _ = { label: 0, sent: function() { if (t[0] & 1) throw t[1]; return t[1]; }, trys: [], ops: [] }, f, y, t, g = Object.create((typeof Iterator === "function" ? Iterator : Object).prototype);
    return g.next = verb(0), g["throw"] = verb(1), g["return"] = verb(2), typeof Symbol === "function" && (g[Symbol.iterator] = function() { return this; }), g;
    function verb(n) { return function (v) { return step([n, v]); }; }
    function step(op) {
        if (f) throw new TypeError("Generator is already executing.");
        while (g && (g = 0, op[0] && (_ = 0)), _) try {
            if (f = 1, y && (t = op[0] & 2 ? y["return"] : op[0] ? y["throw"] || ((t = y["return"]) && t.call(y), 0) : y.next) && !(t = t.call(y, op[1])).done) return t;
            if (y = 0, t) op = [op[0] & 2, t.value];
            switch (op[0]) {
                case 0: case 1: t = op; break;
                case 4: _.label++; return { value: op[1], done: false };
                case 5: _.label++; y = op[1]; op = [0]; continue;
                case 7: op = _.ops.pop(); _.trys.pop(); continue;
                default:
                    if (!(t = _.trys, t = t.length > 0 && t[t.length - 1]) && (op[0] === 6 || op[0] === 2)) { _ = 0; continue; }
                    if (op[0] === 3 && (!t || (op[1] > t[0] && op[1] < t[3]))) { _.label = op[1]; break; }
                    if (op[0] === 6 && _.label < t[1]) { _.label = t[1]; t = op; break; }
                    if (t && _.label < t[2]) { _.label = t[2]; _.ops.push(op); break; }
                    if (t[2]) _.ops.pop();
                    _.trys.pop(); continue;
            }
            op = body.call(thisArg, _);
        } catch (e) { op = [6, e]; y = 0; } finally { f = t = 0; }
        if (op[0] & 5) throw op[1]; return { value: op[0] ? op[1] : void 0, done: true };
    }
};
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.TenantManagementService = void 0;
exports.getTenantManagementService = getTenantManagementService;
var fs_1 = __importDefault(require("fs"));
var path_1 = __importDefault(require("path"));
var child_process_1 = require("child_process");
var util_1 = require("util");
var logger_1 = require("../utils/logger");
var execAsync = (0, util_1.promisify)(child_process_1.exec);
var logger = (0, logger_1.getLogger)("TenantManagementService");
/**
 * TenantManagementService - Manages multi-tenant provisioning
 *
 * Responsibilities:
 *   ✅ Setup new tenant databases (create files + migrations)
 *   ✅ Delete tenant (cleanup databases)
 *   ✅ Check tenant existence
 *   ✅ Load/update tenants configuration
 */
var TenantManagementService = /** @class */ (function () {
    function TenantManagementService() {
        this.tenantsConfigPath = path_1.default.resolve("config/static/tenants.json");
        this.databasesDir = path_1.default.resolve("config/databases");
    }
    /**
     * Check if tenant database exists
     */
    TenantManagementService.prototype.databaseExists = function (tenantId) {
        return __awaiter(this, void 0, void 0, function () {
            var config, dbPath;
            return __generator(this, function (_a) {
                try {
                    config = this.loadTenantsConfig();
                    if (!config[tenantId]) {
                        return [2 /*return*/, false];
                    }
                    dbPath = path_1.default.resolve(config[tenantId].dbPath);
                    return [2 /*return*/, fs_1.default.existsSync(dbPath)];
                }
                catch (error) {
                    logger.error("Failed to check database existence: ".concat(error));
                    return [2 /*return*/, false];
                }
                return [2 /*return*/];
            });
        });
    };
    /**
     * Setup new tenant with Prisma database
     * Creates database file and runs migrations
     */
    TenantManagementService.prototype.setupTenantDatabase = function (tenantId, shopName) {
        return __awaiter(this, void 0, void 0, function () {
            var config, dbFileName, dbPath, error_1;
            return __generator(this, function (_a) {
                switch (_a.label) {
                    case 0:
                        _a.trys.push([0, 2, , 3]);
                        // 1. Validate tenant ID
                        if (!tenantId || tenantId.length === 0) {
                            throw new Error("Invalid tenant ID");
                        }
                        config = this.loadTenantsConfig();
                        if (config[tenantId]) {
                            throw new Error("Tenant ".concat(tenantId, " already exists"));
                        }
                        dbFileName = "".concat(tenantId, "_bertigamart.db");
                        dbPath = path_1.default.join(this.databasesDir, dbFileName);
                        logger.info("\uD83D\uDE80 Setting up database for tenant: ".concat(tenantId));
                        // 4. Create directory if needed
                        if (!fs_1.default.existsSync(this.databasesDir)) {
                            fs_1.default.mkdirSync(this.databasesDir, { recursive: true });
                            logger.info("\uD83D\uDCC1 Created databases directory");
                        }
                        // 5. Create empty database file
                        if (!fs_1.default.existsSync(dbPath)) {
                            fs_1.default.writeFileSync(dbPath, "");
                            logger.info("\uD83D\uDCC4 Created database file: ".concat(dbPath));
                        }
                        // 6. Run Prisma migrations
                        return [4 /*yield*/, this.runMigrations(dbPath)];
                    case 1:
                        // 6. Run Prisma migrations
                        _a.sent();
                        // 7. Update tenants.json
                        this.addTenantToConfig(tenantId, dbPath, shopName);
                        logger.info("\u2705 Tenant ".concat(tenantId, " database setup complete"));
                        return [3 /*break*/, 3];
                    case 2:
                        error_1 = _a.sent();
                        logger.error("\u274C Failed to setup tenant database: ".concat(error_1.message));
                        throw error_1;
                    case 3: return [2 /*return*/];
                }
            });
        });
    };
    /**
     * Setup job database for tenant
     * Auto-called by getTenantJobDatabase if doesn't exist
     */
    TenantManagementService.prototype.setupJobDatabase = function (tenantId) {
        var jobDbPath = path_1.default.join(this.databasesDir, "".concat(tenantId, "_jobs.db"));
        if (!fs_1.default.existsSync(jobDbPath)) {
            fs_1.default.writeFileSync(jobDbPath, "");
            logger.info("\uD83D\uDCC4 Created job database: ".concat(jobDbPath));
        }
    };
    /**
     * Delete tenant completely (cleanup both databases)
     */
    TenantManagementService.prototype.deleteTenant = function (tenantId) {
        return __awaiter(this, void 0, void 0, function () {
            var config, dbPath, jobDbPath;
            return __generator(this, function (_a) {
                try {
                    config = this.loadTenantsConfig();
                    if (!config[tenantId]) {
                        throw new Error("Tenant ".concat(tenantId, " not found"));
                    }
                    logger.info("\uD83D\uDDD1\uFE0F Deleting tenant: ".concat(tenantId));
                    dbPath = path_1.default.resolve(config[tenantId].dbPath);
                    if (fs_1.default.existsSync(dbPath)) {
                        fs_1.default.unlinkSync(dbPath);
                        logger.info("\uD83D\uDDD1\uFE0F Deleted Prisma database: ".concat(dbPath));
                    }
                    jobDbPath = path_1.default.join(this.databasesDir, "".concat(tenantId, "_jobs.db"));
                    if (fs_1.default.existsSync(jobDbPath)) {
                        fs_1.default.unlinkSync(jobDbPath);
                        logger.info("\uD83D\uDDD1\uFE0F Deleted job database: ".concat(jobDbPath));
                    }
                    // 3. Delete from tenants.json
                    this.removeTenantFromConfig(tenantId);
                    logger.info("\u2705 Tenant ".concat(tenantId, " deleted successfully"));
                }
                catch (error) {
                    logger.error("\u274C Failed to delete tenant: ".concat(error.message));
                    throw error;
                }
                return [2 /*return*/];
            });
        });
    };
    /**
     * Get all tenant IDs
     */
    TenantManagementService.prototype.getAllTenantIds = function () {
        try {
            var config = this.loadTenantsConfig();
            return Object.keys(config);
        }
        catch (error) {
            logger.error("Failed to get tenant IDs: ".concat(error));
            return []; // Return empty array, no hardcoded fallback
        }
    };
    /**
     * Get all tenants configuration
     */
    TenantManagementService.prototype.getAllTenants = function () {
        try {
            return this.loadTenantsConfig();
        }
        catch (error) {
            logger.error("Failed to load tenants config: ".concat(error));
            return {};
        }
    };
    /**
     * Load tenants configuration from JSON
     * No hardcoded fallback - returns empty config if not found
     */
    TenantManagementService.prototype.loadTenantsConfig = function () {
        try {
            if (!fs_1.default.existsSync(this.tenantsConfigPath)) {
                logger.warn("Tenants config not found at ".concat(this.tenantsConfigPath));
                return {};
            }
            var content = fs_1.default.readFileSync(this.tenantsConfigPath, "utf-8");
            return JSON.parse(content);
        }
        catch (error) {
            logger.error("Failed to load tenants config: ".concat(error));
            return {};
        }
    };
    /**
     * Add tenant to configuration
     */
    TenantManagementService.prototype.addTenantToConfig = function (tenantId, dbPath, shopName) {
        try {
            var config = this.loadTenantsConfig();
            config[tenantId] = {
                dbPath: dbPath,
                shopName: shopName,
            };
            fs_1.default.writeFileSync(this.tenantsConfigPath, JSON.stringify(config, null, 2));
            logger.info("\u2705 Updated tenants.json with ".concat(tenantId));
        }
        catch (error) {
            logger.error("Failed to update tenants config: ".concat(error));
            throw error;
        }
    };
    /**
     * Remove tenant from configuration
     */
    TenantManagementService.prototype.removeTenantFromConfig = function (tenantId) {
        try {
            var config = this.loadTenantsConfig();
            delete config[tenantId];
            fs_1.default.writeFileSync(this.tenantsConfigPath, JSON.stringify(config, null, 2));
            logger.info("\u2705 Removed ".concat(tenantId, " from tenants.json"));
        }
        catch (error) {
            logger.error("Failed to update tenants config: ".concat(error));
            throw error;
        }
    };
    /**
     * Run Prisma migrations for specific database
     */
    TenantManagementService.prototype.runMigrations = function (dbPath) {
        return __awaiter(this, void 0, void 0, function () {
            var env, stdout, error_2;
            return __generator(this, function (_a) {
                switch (_a.label) {
                    case 0:
                        _a.trys.push([0, 2, , 3]);
                        logger.info("\uD83D\uDD27 Running migrations for ".concat(dbPath, "..."));
                        env = __assign(__assign({}, process.env), { DATABASE_URL: "file:".concat(dbPath) });
                        return [4 /*yield*/, execAsync("npx prisma db push --skip-generate", {
                                env: env,
                                cwd: process.cwd(),
                            })];
                    case 1:
                        stdout = (_a.sent()).stdout;
                        logger.info("\u2705 Migrations completed for ".concat(dbPath));
                        if (stdout)
                            logger.info(stdout);
                        return [3 /*break*/, 3];
                    case 2:
                        error_2 = _a.sent();
                        logger.error("\u274C Migration failed: ".concat(error_2.message));
                        throw error_2;
                    case 3: return [2 /*return*/];
                }
            });
        });
    };
    return TenantManagementService;
}());
exports.TenantManagementService = TenantManagementService;
function getTenantManagementService() {
    return new TenantManagementService();
}
