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

  const hasOpenPorts = (device: Device) => {
    return device.Ports?.some((port) => port.state === 'open');
  };

  const hasFilteredPorts = (device: Device) => {
    return device.Ports?.some((port) => port.state === 'filtered');
  };

  const getDeviceOpacity = (device: Device) => {
    switch (device.Status) {
      case 'idle':
        return 0.8; // Lightly dimmed
      case 'offline':
        return 0.6; // More dimmed
      default:
        return 1; // No dimming
    }
  };

  return (
    <div className="device-container mt-5 d-flex align-items-start">
      <h6 className="text-success d-block w-100">[ DEVICES ]</h6>
      {devices && devices.map((device, index) => (
        <button
          key={index}
          type="button"
          className={`device-box-btn bg-very-dark d-block p-2 me-2 mb-2 w-20-fit text-decoration-none text-start ${device.IPv4 === localDevice?.IPv4 ? 'text-primary' : 'text-success'} border-0`}
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
            <span className="" style={{ fontSize: 24, fontWeight: 600 }}>{device.IPv4}</span><br />
            <span>{device.MAC}</span>
            {device.IPv4 === localDevice?.IPv4 && <span className="f-xs">This device</span>}
          </div>
        </button>
      ))}

      <DeviceModal device={selectedDevice} onClose={() => setSelectedDevice(null)} />
    </div>
  );
};

export default Devices;
