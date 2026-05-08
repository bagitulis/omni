import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { message } from "@/components/AntStaticApi";
import {
  getMonitorData,
  getAutoFunctions,
  cancelJob,
  forceCancelJob,
  clearHistory,
  updateAutoFunction,
  createAutoFunction,
  deleteAutoFunction,
  enableAutoFunction,
  disableAutoFunction,
  runAutoFunction,
  cancelScheduled,
} from "@/api/scriptMonitor";
import { AutoFunctionConfig } from "@/types/scriptMonitor";
import { useAuthStore } from "@/stores/authStore";

export function useScriptMonitor() {
  const queryClient = useQueryClient();
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

  const monitorQuery = useQuery({
    queryKey: ["script-monitor"],
    queryFn: getMonitorData,
    refetchInterval: 5000,
    enabled: isAuthenticated, // Wait for auth before querying
  });

  const autoFunctionsQuery = useQuery({
    queryKey: ["auto-functions"],
    queryFn: getAutoFunctions,
    enabled: isAuthenticated, // Wait for auth before querying
  });

  const cancelJobMutation = useMutation({
    mutationFn: cancelJob,
    onSuccess: () => {
      message.success("Job cancelled");
      queryClient.invalidateQueries({ queryKey: ["script-monitor"] });
    },
    onError: (e: Error) => message.error(e.message),
  });

  const forceCancelJobMutation = useMutation({
    mutationFn: forceCancelJob,
    onSuccess: () => {
      message.success("Job force cancelled");
      queryClient.invalidateQueries({ queryKey: ["script-monitor"] });
    },
    onError: (e: Error) => message.error(e.message),
  });

  const clearHistoryMutation = useMutation({
    mutationFn: clearHistory,
    onSuccess: () => {
      message.success("History cleared");
      queryClient.invalidateQueries({ queryKey: ["script-monitor"] });
    },
    onError: (e: Error) => message.error(e.message),
  });

  const enableAutoFunctionMutation = useMutation({
    mutationFn: enableAutoFunction,
    onSuccess: () => {
      message.success("Auto-function enabled");
      queryClient.invalidateQueries({ queryKey: ["auto-functions"] });
    },
    onError: (e: Error) => message.error(e.message),
  });

  const disableAutoFunctionMutation = useMutation({
    mutationFn: disableAutoFunction,
    onSuccess: () => {
      message.success("Auto-function disabled");
      queryClient.invalidateQueries({ queryKey: ["auto-functions"] });
    },
    onError: (e: Error) => message.error(e.message),
  });

  const updateAutoFunctionMutation = useMutation({
    mutationFn: ({
      name,
      config,
    }: {
      name: string;
      config: Partial<AutoFunctionConfig>;
    }) => updateAutoFunction(name, config),
    onSuccess: () => {
      message.success("Configuration updated");
      queryClient.invalidateQueries({ queryKey: ["auto-functions"] });
    },
    onError: (e: Error) => message.error(e.message),
  });

  const createAutoFunctionMutation = useMutation({
    mutationFn: createAutoFunction,
    onSuccess: () => {
      message.success("Auto-function created");
      queryClient.invalidateQueries({ queryKey: ["auto-functions"] });
    },
    onError: (e: Error) => message.error(e.message),
  });

  const deleteAutoFunctionMutation = useMutation({
    mutationFn: deleteAutoFunction,
    onSuccess: () => {
      message.success("Auto-function deleted");
      queryClient.invalidateQueries({ queryKey: ["auto-functions"] });
    },
    onError: (e: Error) => message.error(e.message),
  });

  const cancelScheduledMutation = useMutation({
    mutationFn: cancelScheduled,
    onSuccess: () => {
      message.success("Scheduled execution cancelled");
      queryClient.invalidateQueries({ queryKey: ["auto-functions"] });
    },
    onError: (e: Error) => message.error(e.message),
  });

  const runAutoFunctionMutation = useMutation({
    mutationFn: runAutoFunction,
    onSuccess: () => {
      message.success("Auto-function execution started");
      queryClient.invalidateQueries({ queryKey: ["script-monitor"] });
      queryClient.invalidateQueries({ queryKey: ["auto-functions"] });
    },
    onError: (e: Error) => message.error(e.message),
  });

  return {
    monitorData: monitorQuery.data,
    isLoadingMonitor: monitorQuery.isLoading,
    autoFunctions: autoFunctionsQuery.data,
    isLoadingAutoFunctions: autoFunctionsQuery.isLoading,

    cancelJob: cancelJobMutation.mutate,
    forceCancelJob: forceCancelJobMutation.mutate,
    clearHistory: clearHistoryMutation.mutate,
    enableAutoFunction: enableAutoFunctionMutation.mutate,
    disableAutoFunction: disableAutoFunctionMutation.mutate,
    runAutoFunction: runAutoFunctionMutation.mutate,
    updateAutoFunction: updateAutoFunctionMutation.mutate,
    createAutoFunction: createAutoFunctionMutation.mutate,
    deleteAutoFunction: deleteAutoFunctionMutation.mutate,
    cancelScheduled: cancelScheduledMutation.mutate,
  };
}
