import React from 'react';
import useEventLogs from '../../hooks/useEventLogs'; // Adjust path as needed
import { EventLog } from '../../models/event-log.model';

const EventLogs = () => {
  const eventLogs: EventLog[] = useEventLogs();

  return (
    <div className="mt-5 col-4">
      <h6 className="text-success d-block w-100">[ EVENT LOG ]</h6>
      <table className="table table-dark table-sm table-compact border-dark border-bottom text-success" style={{ fontSize: '13px' }}>
        <tbody>
          {eventLogs.map((log: EventLog, index: React.Key | null | undefined) => (
            <tr key={index}>
              <td className='bg-transparent text-success px-3'>{log.Type}</td>
              <td className='bg-transparent text-success px-3'>{log.Description}</td>
              <td className="bg-transparent text-success px-3 text-end">{new Date(log.CreatedAt).toLocaleString()}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

export default EventLogs;
