// DeviceList.tsx
import React, { useState } from 'react';
import DeviceModal from '../DeviceModal/DeviceModal'; // Adjust the import path as necessary
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

  const getOpenPortsAmount = (device: Device) => {
    return device.Ports?.filter((port) => port.state === 'open').length || 0;
  };

  const getClosedPortsAmount = (device: Device) => {
    return device.Ports?.filter((port) => port.state === 'closed').length || 0;
  };

  return (
    <div className="device-container mt-5 d-flex align-items-start">
      <h6 className="text-success d-block w-100">[ DEVICES ]</h6>
      {devices.map((device, index) => (
        <button
          key={index}
          type="button"
          className={`device-box-btn bg-very-dark d-block p-2 me-2 mb-2 w-20-fit text-decoration-none text-start ${device.IPv4 === localDevice?.IPv4 ? 'text-primary' : 'text-success'} border-0`}          onClick={() => setSelectedDevice(device)}
        >
          <div className="d-flex justify-content-between align-items-baseline">
            {hasOpenPorts(device) && (
              <span className="fw-bold">
                <small>
                  <i className="fas fa-exclamation-triangle fa-xs"></i> {getOpenPortsAmount(device)}
                </small>
              </span>
            )}
            <span>
              <small>{getClosedPortsAmount(device)}</small>
            </span>
          </div>
          <div className="mt-3">
            <span className="fw-bold">{device.IPv4}</span><br />
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
