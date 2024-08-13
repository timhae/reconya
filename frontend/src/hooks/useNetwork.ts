import { useEffect, useState } from "react";
import { fetchNetwork } from "../api/axiosConfig";
import { Network } from "../models/network.model";

const useNetwork = () => {
  const [network, setNetwork] = useState<Network | undefined>();
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const getNetwork = async () => {
      try {
        const data = await fetchNetwork();
        setNetwork(data);
      } catch (error: any) {
        setError(error);
      } finally {
        setIsLoading(false);
      }
    };

    getNetwork();

    const interval = setInterval(getNetwork, 10000);

    return () => clearInterval(interval);
  },[]);

  return { network, isLoading, error };
};

export default useNetwork;
