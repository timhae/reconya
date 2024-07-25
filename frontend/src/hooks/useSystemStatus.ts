import { useEffect, useState } from "react";
import { fetchSystemStatus } from "../api/axiosConfig";
import { SystemStatus } from "../models/system-status.model";

// This is a simplistic extension. You might want to split this into multiple hooks or handle state more granularly.
const useSystemStatus = () => {
  const [systemStatus, setSystemStatus] = useState<SystemStatus | undefined>();
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const getSystemStatus = async () => {
      try {
        // setIsLoading(true);
        const data = await fetchSystemStatus();
        setSystemStatus({ LocalDevice: data.LocalDevice });
      } catch (error: any) {
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
