import React from 'react';
import { Device } from '../../models/device.model';

interface Props {
  devices: Device[];
  localDevice?: Device;
}

const getDeviceContainerCssClasses = (device: Device, localDevice?: Device): string => {
  if (device.Status === 'online' && device.IPv4 !== localDevice?.IPv4) {
    return 'border border-success text-success';
  } else if (device.IPv4 === localDevice?.IPv4) {
    return 'border border-primary text-primary';
  } else {
    if (device.Status === 'offline') {
      return 'border border-dark text-dark';
    } else if (device.Status === 'idle') {
      return 'border border-success text-success opacity-50';
    }
    return 'border border-dark';
  }
};

const NetworkMap: React.FC<Props> = ({ devices, localDevice }) => {
  return (
    <div className="device-container mt-4 d-flex align-items-start">
      <h6 className="text-success d-block w-100">[ NETWORK MAP ]</h6>
      {devices.map((device, index) => (
        <button
          key={index}
          className={`device-mini-box bg-very-dark d-block ps-2 pe-0 ${getDeviceContainerCssClasses(device, localDevice)}`}
          title={device.IPv4}
        >
          {/* Placeholder for icon */}
        </button>
      ))}
    </div>
  );
};

export default NetworkMap;
