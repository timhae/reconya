import React from 'react';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faExclamationCircle, faLock } from '@fortawesome/free-solid-svg-icons';
import { Device } from '../../models/device.model';

interface DeviceListProps {
  devices: Device[];
  localDevice?: Device;
}

const DeviceList: React.FC<DeviceListProps> = ({ devices, localDevice }) => {
  const formatDate = (dateString: string | undefined) => {
    if (!dateString || dateString.startsWith("0001-01-01")) return "Unknown";
    const date = new Date(dateString);
    return `${date.toLocaleDateString('en-US', {
      year: '2-digit',
      month: 'numeric',
      day: 'numeric'
    })} ${date.toLocaleTimeString('en-US', {
      hour: '2-digit',
      minute: '2-digit'
    })}`;
  };

  const getOpenPortsAmount = (device: Device) => {
    return device.Ports?.filter((port) => port.state === 'open').length || 0;
  };

  const getFilteredPortsAmount = (device: Device) => {
    return device.Ports?.filter((port) => port.state === 'filtered').length || 0;
  };

  const getDeviceOpacity = (device: Device) => {
    switch (device.Status) {
      case 'idle':
        return 0.8; 
      case 'offline':
      case 'unknown':
        return 0.6;
      default:
        return 1; 
    }
  };

  return (
    <div className="mt-5">
      <h6 className="text-success d-block w-100">[ ONLINE DEVICE LIST ]</h6>
      <table className="table-dark table-sm table-compact text-success w-100 border border-dark">
        <thead className="border-bottom border-dark bg-success text-dark">
          <tr>
            <th className="px-3 py-1">Hostname</th>
            <th className="px-3 py-1">IPv4</th>
            <th className="px-3 py-1">MAC</th>
            <th className="px-3 py-1">Vendor</th>
            <th className="px-3 py-1 text-center" style={{ width: '120px' }}>Open Ports</th>
            <th className="px-3 py-1 text-center" style={{ width: '130px' }}>Filtered Ports</th>
            <th className="px-3 py-1">Last Seen Online</th>
          </tr>
        </thead>
        <tbody>
          {devices.map((device, index) => (
            <tr
              key={index}
              className={`border-bottom border-dark ${device.IPv4 === localDevice?.IPv4 ? 'text-primary' : ''}`}
              style={{ opacity: getDeviceOpacity(device) }}
            >
              <td className="px-3 py-1">{device.Hostname || 'Unknown'}</td>
              <td className="px-3 py-1">{device.IPv4}</td>
              <td className="px-3 py-1">{device.MAC || 'N/A'}</td>
              <td className="px-3 py-1">{device.Vendor || 'Unknown'}</td>
              <td className="px-3 py-1 text-center">
                {getOpenPortsAmount(device) > 0 ? (
                  <span className="text-danger">
                    <FontAwesomeIcon icon={faExclamationCircle} className="me-1" />
                    {getOpenPortsAmount(device)}
                  </span>
                ) : (
                  "-"
                )}
              </td>
              <td className="px-3 py-1 text-center">
                {getFilteredPortsAmount(device) > 0 ? (
                  <span className="text-warning">
                    <FontAwesomeIcon icon={faLock} className="me-1" />
                    {getFilteredPortsAmount(device)}
                  </span>
                ) : (
                  "-"
                )}
              </td>
              <td className="px-3 py-1">{formatDate(device.LastSeenOnlineAt)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

export default DeviceList;
