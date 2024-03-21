import React from 'react';
import { Device } from '../../models/device.model';

interface DeviceListProps {
  devices: Device[];
  localDevice?: Device; // Add localDevice to the interface
}

const DeviceList: React.FC<DeviceListProps> = ({ devices, localDevice }) => {
  const formatDate = (dateString: string | undefined) => {
    if (!dateString || dateString.startsWith("0001-01-01")) return "Unknown";
    return new Date(dateString).toLocaleDateString("en-US", {
      year: 'numeric', month: 'long', day: 'numeric',
      hour: '2-digit', minute: '2-digit', second: '2-digit',
    });
  };

  return (
    <div className="mt-5">
      <h6 className="text-success d-block w-100">[ ONLINE DEVICE LIST ]</h6>
      <table className="table-dark table-sm table-compact text-success w-100">
        <thead>
          <tr>
            <th>Hostname</th>
            <th>IPv4</th>
            <th>MAC</th>
            <th>Vendor</th>
            <th>Portscan</th>
            <th>Last Seen Online</th>
          </tr>
        </thead>
        <tbody>
          {devices.map((device, index) => (
            <tr key={index} className={device.IPv4 === localDevice?.IPv4 ? 'text-primary' : ''}>
              <td>{device.Hostname || 'Unknown'}</td>
              <td>{device.IPv4}</td>
              <td>{device.MAC || 'N/A'}</td>
              <td>{device.Vendor || 'Unknown'}</td>
              <td>
                {device.Ports ? `${device.Ports.filter(p => p.state === 'open').length} / ${device.Ports.filter(p => p.state === 'closed').length}` : 'N/A'}
              </td>
              <td>{formatDate(device.LastSeenOnlineAt)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

export default DeviceList;
