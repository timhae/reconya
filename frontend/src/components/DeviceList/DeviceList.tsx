import React from 'react';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faExclamationCircle, faLock } from '@fortawesome/free-solid-svg-icons';
import { Device } from '../../models/device.model';

interface DeviceListProps {
  devices: Device[];
  localDevice?: Device;
}

const DeviceList: React.FC<DeviceListProps> = ({ devices, localDevice }) => {
  // Helper functions to normalize property access
  const getDeviceID = (device: Device) => device.id || device.ID || '';
  const getDeviceIPv4 = (device: Device) => device.ipv4 || device.IPv4 || '';
  const getDeviceMAC = (device: Device) => device.mac || device.MAC;
  const getDeviceVendor = (device: Device) => device.vendor || device.Vendor;
  const getDeviceHostname = (device: Device) => device.hostname || device.Hostname;
  const getDeviceStatus = (device: Device) => device.status || device.Status;
  const getDevicePorts = (device: Device) => device.ports || device.Ports || [];
  const getDeviceLastSeen = (device: Device) => device.last_seen_online_at || device.LastSeenOnlineAt;

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
    return getDevicePorts(device)?.filter((port) => port.state === 'open').length || 0;
  };

  const getFilteredPortsAmount = (device: Device) => {
    return getDevicePorts(device)?.filter((port) => port.state === 'filtered').length || 0;
  };

  const getDeviceOpacity = (device: Device) => {
    switch (getDeviceStatus(device)) {
      case 'idle':
        return 0.8; 
      case 'offline':
      case 'unknown':
        return 0.6;
      default:
        return 1; 
    }
  };

  // Remove duplicate IP addresses by creating a map with unique IPs
  const uniqueDevicesByIP = devices.reduce((acc, device) => {
    const ipv4 = getDeviceIPv4(device);
    // Prefer devices with more information (MAC, hostname, ports)
    if (!acc[ipv4] || 
        (!getDeviceMAC(acc[ipv4]) && getDeviceMAC(device)) || 
        (!getDeviceHostname(acc[ipv4]) && getDeviceHostname(device)) ||
        (!getDevicePorts(acc[ipv4])?.length && getDevicePorts(device)?.length)) {
      acc[ipv4] = device;
    }
    return acc;
  }, {} as Record<string, Device>);

  // Convert back to array
  const uniqueDevices = Object.values(uniqueDevicesByIP);

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
          {uniqueDevices.map((device) => {
            const localIpv4 = localDevice ? getDeviceIPv4(localDevice) : '';
            
            return (
              <tr
                key={getDeviceID(device)}
                className={`border-bottom border-dark ${getDeviceIPv4(device) === localIpv4 ? 'text-primary' : ''}`}
                style={{ opacity: getDeviceOpacity(device) }}
              >
                <td className="px-3 py-1">{getDeviceHostname(device) || 'Unknown'}</td>
                <td className="px-3 py-1">{getDeviceIPv4(device)}</td>
                <td className="px-3 py-1">{getDeviceMAC(device) || 'N/A'}</td>
                <td className="px-3 py-1">{getDeviceVendor(device) || 'Unknown'}</td>
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
                <td className="px-3 py-1">{formatDate(getDeviceLastSeen(device))}</td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
};

export default DeviceList;
