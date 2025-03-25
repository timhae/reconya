import { useEffect, useState } from "react";
import { fetchSystemStatus } from "../api/axiosConfig";
import { SystemStatus } from "../models/systemStatus.model";

const useSystemStatus = () => {
  const [systemStatus, setSystemStatus] = useState<SystemStatus | undefined>();
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const getSystemStatus = async () => {
      try {
        // setIsLoading(true);
        const data = await fetchSystemStatus();
        console.log("System status response:", data);
        setSystemStatus(data);
      } catch (error: any) {
        console.error("Error fetching system status:", error);
        setError(error);
      } finally {
        setIsLoading(false);
      }
    };

    getSystemStatus();

    const interval = setInterval(getSystemStatus, 10000);

    return () => clearInterval(interval);
  }, []);

  return { systemStatus, isLoading, error };
};

export default useSystemStatus;
