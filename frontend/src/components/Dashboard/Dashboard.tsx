// src/components/Dashboard/Dashboard.tsx
import React from 'react';
import useAuth from '../../hooks/useAuth';
import useDevices from '../../hooks/useDevices';
import useSystemStatus from '../../hooks/useSystemStatus';
import NetworkMap from '../NetworkMap/NetworkMap';
import Devices from '../Devices/Devices';
import DeviceList from '../DeviceList/DeviceList';
import EventLogs from '../EventLog/EventLogs';
import SystemStatusComponent from '../SystemStatus/SystemStatus';
import useNetwork from '../../hooks/useNetwork';

const Dashboard: React.FC = () => {
  const isAuthenticated = useAuth();

  const { devices, isLoading: devicesLoading, error: devicesError } = useDevices();
  const { systemStatus, isLoading: systemStatusLoading, error: systemStatusError } = useSystemStatus();
  const { network } = useNetwork();

  const localDevice = systemStatus?.LocalDevice;

  const isLoading = devicesLoading || systemStatusLoading;
  const error = devicesError || systemStatusError;

  if (!isAuthenticated) return null;
  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;

  return (
    <div className="container-fluid">
      <div className="row mt-1">
        <div className="col-md-9">
          <NetworkMap devices={devices} localDevice={localDevice} />
          <Devices devices={devices} localDevice={localDevice} />
        </div>
        <div className="col-md-3 d-flex flex-column">
          <div className="mb-4">
            <SystemStatusComponent systemStatus={systemStatus} network={network} />
          </div>
          <div className="flex-grow-1">
            <EventLogs />
          </div>
        </div>
      </div>
      <div className="mt-4">
        <DeviceList devices={devices} localDevice={localDevice} />
      </div>
    </div>
  );
};

export default Dashboard;
