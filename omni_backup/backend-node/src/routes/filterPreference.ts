import { Router } from "express";
import filterPreferenceController from "../controllers/filterPreferenceController";

const router = Router();

/**
 * Filter Preference Routes
 * Save/Load filter preferences (column visibility, column filters, search query)
 */

// Save filter preference (POST directly to /)
router.post(
  "/",
  filterPreferenceController.saveFilterPreference.bind(
    filterPreferenceController
  )
);

// Get filter preference (GET /)
router.get(
  "/",
  filterPreferenceController.getFilterPreference.bind(
    filterPreferenceController
  )
);

// Delete filter preference (DELETE /)
router.delete(
  "/",
  filterPreferenceController.deleteFilterPreference.bind(
    filterPreferenceController
  )
);

export default router;
