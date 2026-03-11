import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { message } from "@/components/AntStaticApi";
import {
  getSavedLinks,
  getDetailedSettings,
  validateLink,
  saveLinks,
  updateSettings,
} from "@/api/googleSheets";

/**
 * Hook for fetching Google Sheets links
 */
export function useGoogleSheetsLinks() {
  return useQuery({
    queryKey: ["google-sheets-links"],
    queryFn: getSavedLinks,
  });
}

/**
 * Hook for fetching detailed Google Sheets settings
 */
export function useGoogleSheetsDetails() {
  return useQuery({
    queryKey: ["google-sheets-details"],
    queryFn: getDetailedSettings,
  });
}

/**
 * Hook for validating Google Sheets link
 */
export function useValidateLink() {
  return useMutation({
    mutationFn: validateLink,
    onSuccess: () => {
      message.success("Link validated successfully");
    },
    onError: (e: Error) => {
      message.error(e.message);
    },
  });
}

/**
 * Hook for saving Google Sheets links
 */
export function useSaveLinks() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: saveLinks,
    onSuccess: () => {
      message.success("Links saved successfully");
      queryClient.invalidateQueries({ queryKey: ["google-sheets-links"] });
    },
    onError: (e: Error) => {
      message.error(e.message);
    },
  });
}

/**
 * Hook for updating detailed Google Sheets settings
 */
export function useUpdateSettings() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: updateSettings,
    onSuccess: () => {
      message.success("Settings updated successfully");
      queryClient.invalidateQueries({ queryKey: ["google-sheets-links"] });
      queryClient.invalidateQueries({ queryKey: ["google-sheets-details"] });
    },
    onError: (e: Error) => {
      message.error(e.message);
    },
  });
}
