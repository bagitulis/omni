"use strict";
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
exports.getLogger = getLogger;
exports.closeLoggers = closeLoggers;
var winston_1 = __importDefault(require("winston"));
var loggers = new Map();
/**
 * Get or create a logger instance
 * Singleton pattern to avoid duplicate logger instances
 */
function getLogger(label) {
    if (loggers.has(label)) {
        return loggers.get(label);
    }
    var logger = winston_1.default.createLogger({
        level: process.env.LOG_LEVEL || "info",
        format: winston_1.default.format.combine(winston_1.default.format.timestamp({ format: "YYYY-MM-DD HH:mm:ss" }), winston_1.default.format.printf(function (_a) {
            var timestamp = _a.timestamp, level = _a.level, message = _a.message, label = _a.label;
            var emoji = getEmojiForLevel(level);
            return "".concat(emoji, " [").concat(timestamp, "] [").concat(label, "] ").concat(message);
        })),
        defaultMeta: { label: label },
        transports: [
            new winston_1.default.transports.Console(),
            new winston_1.default.transports.File({
                filename: "logs/error.log",
                level: "error",
                maxsize: 5242880, // 5MB
                maxFiles: 5,
            }),
            new winston_1.default.transports.File({
                filename: "logs/combined.log",
                maxsize: 5242880, // 5MB
                maxFiles: 10,
            }),
        ],
    });
    loggers.set(label, logger);
    return logger;
}
/**
 * Get emoji for log level
 */
function getEmojiForLevel(level) {
    switch (level.toLowerCase()) {
        case "error":
            return "❌";
        case "warn":
            return "⚠️";
        case "info":
            return "ℹ️";
        case "debug":
            return "🔧";
        default:
            return "📝";
    }
}
/**
 * Close all loggers gracefully
 */
function closeLoggers() {
    return __awaiter(this, void 0, void 0, function () {
        var _loop_1, _i, loggers_1, _a, _label, logger;
        return __generator(this, function (_b) {
            switch (_b.label) {
                case 0:
                    _loop_1 = function (_label, logger) {
                        return __generator(this, function (_c) {
                            switch (_c.label) {
                                case 0:
                                    if (!(logger && typeof logger.close === "function")) return [3 /*break*/, 2];
                                    return [4 /*yield*/, new Promise(function (resolve) {
                                            logger.close(function () {
                                                resolve();
                                            });
                                        })];
                                case 1:
                                    _c.sent();
                                    _c.label = 2;
                                case 2: return [2 /*return*/];
                            }
                        });
                    };
                    _i = 0, loggers_1 = loggers;
                    _b.label = 1;
                case 1:
                    if (!(_i < loggers_1.length)) return [3 /*break*/, 4];
                    _a = loggers_1[_i], _label = _a[0], logger = _a[1];
                    return [5 /*yield**/, _loop_1(_label, logger)];
                case 2:
                    _b.sent();
                    _b.label = 3;
                case 3:
                    _i++;
                    return [3 /*break*/, 1];
                case 4:
                    loggers.clear();
                    return [2 /*return*/];
            }
        });
    });
}
