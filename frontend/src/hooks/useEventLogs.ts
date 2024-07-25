import { useEffect, useState } from 'react';
import { EventLog } from '../models/event-log.model';
import { fetchEventLogs } from '../api/axiosConfig';

const useEventLogs = () => {
  const [eventLogs, setEventLogs] = useState<EventLog[]>([]); 

  useEffect(() => {
    const fetchLogs = async () => {
      try {
        const logs = await fetchEventLogs();
        setEventLogs(logs);
      } catch (error) {
        console.error("Error fetching event logs:", error);
      }
    };

    fetchLogs();

    const interval = setInterval(fetchLogs, 3000);

    return () => clearInterval(interval);
  }, []);

  return eventLogs;
};

export default useEventLogs;
