// src/components/Dashboard/Dashboard.tsx
import React from 'react';
import useAuth from '../../hooks/useAuth';
import useDevices from '../../hooks/useDevices';
import useSystemStatus from '../../hooks/useSystemStatus';
import NetworkMap from '../NetworkMap/NetworkMap';
import Devices from '../Devices/Devices';
import DeviceList from '../DeviceList/DeviceList';
import EventLogs from '../EventLog/EventLogs';

const Dashboard: React.FC = () => {
  const isAuthenticated = useAuth();

  const { devices, isLoading: devicesLoading, error: devicesError } = useDevices();
  const { systemStatus, isLoading: systemStatusLoading, error: systemStatusError } = useSystemStatus();

  const localDevice = systemStatus?.LocalDevice;

  const isLoading = devicesLoading || systemStatusLoading;
  const error = devicesError || systemStatusError;

  if (!isAuthenticated) return null;
  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;

  return (
    <div>
      <NetworkMap devices={devices} localDevice={localDevice} />
      <EventLogs />
      <Devices devices={devices} localDevice={localDevice} />
      <DeviceList devices={devices} localDevice={localDevice} />
    </div>
  );
};

export default Dashboard;
