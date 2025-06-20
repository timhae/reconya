import React, { useState } from 'react';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faGlobe, faShield, faShieldAlt, faExclamationTriangle, faCamera } from '@fortawesome/free-solid-svg-icons';
import DeviceModal from '../DeviceModal/DeviceModal';
import { Device } from '../../models/device.model';

interface DevicesProps {
  devices: Device[];
  localDevice?: Device;
  onDeviceUpdate?: (updatedDevice: Device) => void;
}

const Devices: React.FC<DevicesProps> = ({ devices, localDevice, onDeviceUpdate }) => {
  const [selectedDevice, setSelectedDevice] = useState<Device | null>(null);

  const handleDeviceUpdate = (updatedDevice: Device) => {
    // Update the selected device with the new data
    setSelectedDevice(updatedDevice);
    // Call the parent's update function
    if (onDeviceUpdate) {
      onDeviceUpdate(updatedDevice);
    }
  };

  // Helper functions to normalize property access
  const getDeviceID = (device: Device) => device.id || device.ID || '';
  const getDeviceIPv4 = (device: Device) => device.ipv4 || device.IPv4 || '';
  const getDeviceMAC = (device: Device) => device.mac || device.MAC;
  const getDeviceHostname = (device: Device) => device.hostname || device.Hostname;
  const getDeviceStatus = (device: Device) => device.status || device.Status;
  const getDevicePorts = (device: Device) => device.ports || device.Ports || [];
  const getDeviceWebServices = (device: Device) => device.web_services || device.WebServices || [];
  const getDeviceType = (device: Device) => device.device_type || device.DeviceType;
  const getDeviceName = (device: Device) => device.name || device.Name;

  const hasOpenPorts = (device: Device) => {
    return getDevicePorts(device).some((port) => port.state === 'open');
  };

  const hasFilteredPorts = (device: Device) => {
    return getDevicePorts(device).some((port) => port.state === 'filtered');
  };

  const hasWebServices = (device: Device) => {
    return getDeviceWebServices(device).length > 0;
  };

  const hasScreenshots = (device: Device) => {
    return getDeviceWebServices(device).some(service => service.screenshot);
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
            className={`device-box-btn bg-very-dark d-block p-2 me-2 mb-2 text-decoration-none text-start ${ipv4 === localIpv4 ? 'text-primary' : 'text-success'} border-0`}
            style={{
              minWidth: 205,
              height: 120,
              opacity: getDeviceOpacity(device),
            }}
            onClick={() => setSelectedDevice(device)}
          >
            <div className="d-flex flex-column h-100">
              <div className="d-flex justify-content-between align-items-start">
                <div className="flex-grow-1">
                  <span className="" style={{ fontSize: 24, fontWeight: 600 }}>{ipv4}</span><br />
                  {/* {mac && <span style={{ fontSize: '0.85rem' }}>{mac}</span>} */}
                  {getDeviceName(device) && (
                    <div
                      style={{
                        width: '160px', // match or slightly less than card minWidth
                        display: 'block',
                        whiteSpace: 'nowrap',
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                        fontSize: '0.85rem',
                        paddingRight: 2
                      }}
                    >
                      <span className="f-xs text-success">{getDeviceName(device)}</span>
                    </div>
                  )}
                  {hostname && <div><span className="f-xs">{hostname}</span></div>}
                </div>
                <div className="d-flex flex-column align-items-end" style={{ gap: '6px' }}>
                  {hasWebServices(device) && (
                    <FontAwesomeIcon 
                      icon={faGlobe} 
                      className="text-success" 
                      style={{ fontSize: '0.7rem' }}
                      title={`${getDeviceWebServices(device).length} web service(s) available`}
                    />
                  )}
                  {hasScreenshots(device) && (
                    <FontAwesomeIcon 
                      icon={faCamera} 
                      className="text-info" 
                      style={{ fontSize: '0.7rem' }}
                      title="Screenshots available"
                    />
                  )}
                  {hasOpenPorts(device) && (
                    <FontAwesomeIcon 
                      icon={faExclamationTriangle} 
                      className="text-danger" 
                      style={{ fontSize: '0.7rem' }}
                      title="Open ports detected"
                    />
                  )}
                  {hasFilteredPorts(device) && !hasOpenPorts(device) && (
                    <FontAwesomeIcon 
                      icon={faShieldAlt} 
                      className="text-warning" 
                      style={{ fontSize: '0.7rem' }}
                      title="Filtered ports detected"
                    />
                  )}
                  {!hasFilteredPorts(device) && !hasOpenPorts(device) && (
                    <FontAwesomeIcon 
                      icon={faShield} 
                      className="text-success" 
                      style={{ fontSize: '0.7rem' }}
                      title="No open ports - secure"
                    />
                  )}
                </div>
              </div>
              <div className="mt-auto">
                {getDeviceType(device) && (
                  <div className="mb-1">
                    <span className="badge bg-dark border border-success text-success" style={{ fontSize: '0.6rem' }}>
                      {getDeviceType(device)?.replace('_', ' ').toUpperCase()}
                    </span>
                  </div>
                )}
                {ipv4 === localIpv4 && <div><span className="f-xs">This device</span></div>}
              </div>
            </div>
          </button>
        );
      })}

      <DeviceModal device={selectedDevice} onClose={() => setSelectedDevice(null)} onDeviceUpdate={handleDeviceUpdate} />
    </div>
  );
};

export default Devices;
