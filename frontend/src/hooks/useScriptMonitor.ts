import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { message } from "antd";
import * as api from "@/api/scriptMonitor";
import { AutoFunctionConfig } from "@/types/scriptMonitor";

export function useScriptMonitor() {
  const queryClient = useQueryClient();

  const monitorQuery = useQuery({
    queryKey: ["script-monitor"],
    queryFn: api.getMonitorData,
    refetchInterval: 2000,
  });

  const autoFunctionsQuery = useQuery({
    queryKey: ["auto-functions"],
    queryFn: api.getAutoFunctions,
  });

  const cancelJobMutation = useMutation({
    mutationFn: api.cancelJob,
    onSuccess: () => {
      message.success("Job cancelled");
      queryClient.invalidateQueries({ queryKey: ["script-monitor"] });
    },
    onError: (e: Error) => message.error(e.message),
  });

  const forceCancelJobMutation = useMutation({
    mutationFn: api.forceCancelJob,
    onSuccess: () => {
      message.success("Job force cancelled");
      queryClient.invalidateQueries({ queryKey: ["script-monitor"] });
    },
    onError: (e: Error) => message.error(e.message),
  });

  const clearHistoryMutation = useMutation({
    mutationFn: api.clearHistory,
    onSuccess: () => {
      message.success("History cleared");
      queryClient.invalidateQueries({ queryKey: ["script-monitor"] });
    },
    onError: (e: Error) => message.error(e.message),
  });

  const enableAutoFunctionMutation = useMutation({
    mutationFn: api.enableAutoFunction,
    onSuccess: () => {
      message.success("Auto-function enabled");
      queryClient.invalidateQueries({ queryKey: ["auto-functions"] });
    },
    onError: (e: Error) => message.error(e.message),
  });

  const disableAutoFunctionMutation = useMutation({
    mutationFn: api.disableAutoFunction,
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
    }) => api.updateAutoFunction(name, config),
    onSuccess: () => {
      message.success("Configuration updated");
      queryClient.invalidateQueries({ queryKey: ["auto-functions"] });
    },
    onError: (e: Error) => message.error(e.message),
  });

  const createAutoFunctionMutation = useMutation({
    mutationFn: api.createAutoFunction,
    onSuccess: () => {
      message.success("Auto-function created");
      queryClient.invalidateQueries({ queryKey: ["auto-functions"] });
    },
    onError: (e: Error) => message.error(e.message),
  });

  const deleteAutoFunctionMutation = useMutation({
    mutationFn: api.deleteAutoFunction,
    onSuccess: () => {
      message.success("Auto-function deleted");
      queryClient.invalidateQueries({ queryKey: ["auto-functions"] });
    },
    onError: (e: Error) => message.error(e.message),
  });

  const cancelScheduledMutation = useMutation({
    mutationFn: api.cancelScheduled,
    onSuccess: () => {
      message.success("Scheduled execution cancelled");
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
    updateAutoFunction: updateAutoFunctionMutation.mutate,
    createAutoFunction: createAutoFunctionMutation.mutate,
    deleteAutoFunction: deleteAutoFunctionMutation.mutate,
    cancelScheduled: cancelScheduledMutation.mutate,
  };
}
