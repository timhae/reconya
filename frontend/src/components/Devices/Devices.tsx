import React, { useState } from 'react';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faCircle } from '@fortawesome/free-solid-svg-icons';
import DeviceModal from '../DeviceModal/DeviceModal';
import { Device } from '../../models/device.model';

interface DevicesProps {
  devices: Device[];
  localDevice?: Device;
}

const Devices: React.FC<DevicesProps> = ({ devices, localDevice }) => {
  const [selectedDevice, setSelectedDevice] = useState<Device | null>(null);

  // Helper functions to normalize property access
  const getDeviceID = (device: Device) => device.id || device.ID || '';
  const getDeviceIPv4 = (device: Device) => device.ipv4 || device.IPv4 || '';
  const getDeviceMAC = (device: Device) => device.mac || device.MAC;
  const getDeviceHostname = (device: Device) => device.hostname || device.Hostname;
  const getDeviceStatus = (device: Device) => device.status || device.Status;
  const getDevicePorts = (device: Device) => device.ports || device.Ports || [];

  const hasOpenPorts = (device: Device) => {
    return getDevicePorts(device).some((port) => port.state === 'open');
  };

  const hasFilteredPorts = (device: Device) => {
    return getDevicePorts(device).some((port) => port.state === 'filtered');
  };

  const getDeviceOpacity = (device: Device) => {
    switch (getDeviceStatus(device)) {
      case 'idle':
        return 0.8; // Lightly dimmed
      case 'offline':
        return 0.6; // More dimmed
      default:
        return 1; // No dimming
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
    <div className="device-container mt-5 d-flex align-items-start">
      <h6 className="text-success d-block w-100">[ DEVICES ]</h6>
      {uniqueDevices.map((device) => {
        const ipv4 = getDeviceIPv4(device);
        const mac = getDeviceMAC(device);
        const hostname = getDeviceHostname(device);
        const localIpv4 = localDevice ? getDeviceIPv4(localDevice) : '';
        
        return (
          <button
            key={getDeviceID(device)}
            type="button"
            className={`device-box-btn bg-very-dark d-block p-2 me-2 mb-2 w-20-fit text-decoration-none text-start ${ipv4 === localIpv4 ? 'text-primary' : 'text-success'} border-0`}
            style={{
              minWidth: 205,
              opacity: getDeviceOpacity(device),
            }}
            onClick={() => setSelectedDevice(device)}
          >
            <div className="d-flex justify-content-end align-items-center">
              {hasOpenPorts(device) && (
                <FontAwesomeIcon 
                  icon={faCircle} 
                  className="text-danger me-1" 
                  style={{ fontSize: '0.4rem' }} 
                />
              )}
              {hasFilteredPorts(device) && (
                <FontAwesomeIcon 
                  icon={faCircle} 
                  className="text-warning" 
                  style={{ fontSize: '0.4rem' }} 
                />
              )}
              {!hasFilteredPorts(device) && !hasOpenPorts(device) && (
                <FontAwesomeIcon 
                  icon={faCircle} 
                  className="text-success" 
                  style={{ fontSize: '0.4rem' }}
                />
              )}
            </div>
            <div className="mt-3">
              <span className="" style={{ fontSize: 24, fontWeight: 600 }}>{ipv4}</span><br />
              <span>{mac}</span>
              {hostname && <div><span className="f-xs">{hostname}</span></div>}
              {ipv4 === localIpv4 && <div><span className="f-xs">This device</span></div>}
            </div>
          </button>
        );
      })}

      <DeviceModal device={selectedDevice} onClose={() => setSelectedDevice(null)} />
    </div>
  );
};

export default Devices;
