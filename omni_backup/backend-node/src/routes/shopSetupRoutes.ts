import { Router, Request, Response } from "express";
import { ShopSetupController } from "../controllers/shopSetupController";

const router = Router();
const controller = new ShopSetupController();

/**
 * Shop Setup Routes
 * For managing global credentials and shop connections
 * Access: Owner, Developer, Admin
 */

// Get global credentials status (masked values)
router.get("/credentials", async (req: Request, res: Response) => {
  await controller.getCredentialsStatus(req, res);
});

// Save global credentials (Owner only)
router.post("/credentials", async (req: Request, res: Response) => {
  await controller.saveCredentials(req, res);
});

// Get current tenant's shop connection status
router.get("/shop-status", async (req: Request, res: Response) => {
  await controller.getShopStatus(req, res);
});

// Disconnect shop from platform
router.delete("/disconnect/:platform", async (req: Request, res: Response) => {
  await controller.disconnectShop(req, res);
});

export default router;
