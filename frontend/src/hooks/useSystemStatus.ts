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
        setSystemStatus({ 
          LocalDevice: data.LocalDevice,
          PublicIP: data.PublicIP
        });
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
