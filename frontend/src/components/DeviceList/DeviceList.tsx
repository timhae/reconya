import React from 'react';
import { Device } from '../../models/device.model';

interface DeviceListProps {
  devices: Device[];
  localDevice?: Device;
}

const DeviceList: React.FC<DeviceListProps> = ({ devices, localDevice }) => {
  const identify = (index: number) => index; // Define identify function if needed

  const hasOpenPorts = (device: Device) => {
    return device.Ports?.filter((port) => port.state === 'open')?? false;
  };

  const getOpenPortsAmount = (device: Device) => {
    return device.Ports?.filter((port) => port.state === 'open')?.length || 0;
  };

  const getClosedPortsAmount = (device: Device) => {
    return device.Ports?.filter((port) => port.state === 'closed')?.length || 0;
  };

  return (
    <div className="device-container mt-5 d-flex align-items-start">
      <h6 className="text-success d-block w-100">[ DEVICES ]</h6>
      {devices.map((device, index) => (
        <button
          key={identify(index)}
          type="button"
          className="device-box-btn bg-very-dark d-block p-2 me-2 mb-2 w-20-fit text-decoration-none text-start text-success border-0"
          data-bs-toggle="modal"
          data-bs-target="#exampleModal"
        >
          <div className="d-flex justify-content-between align-items-baseline">
            <span className="fw-bold">
              {hasOpenPorts(device) && (
                <small>
                  <i className="fas fa-exclamation-triangle fa-xs"></i> {getOpenPortsAmount(device)}
                </small>
              )}
            </span>
            <span className="ps-2">
              <small>{getClosedPortsAmount(device)}</small>
            </span>
          </div>
          <div className="mt-3">
            <span className="fw-bold">{device.IPv4}</span>
            <br />
            <span>{device.MAC}</span>
            {device.IPv4 === localDevice?.IPv4 && <span className="f-xs">This device</span>}
          </div>
        </button>
      ))}
    </div>
  );
};

export default  React.memo(DeviceList);
