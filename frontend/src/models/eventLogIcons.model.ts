import { EEventLogType } from './eventLog.model';
import { faSearch, faExclamationCircle, faClock, faTimesCircle, faGlobe, faNetworkWired, faShieldAlt, faCircleCheck, faSpinner, faFlagCheckered } from '@fortawesome/free-solid-svg-icons';
import { IconDefinition } from '@fortawesome/fontawesome-svg-core';

export const EventLogIcons: Record<EEventLogType, IconDefinition> = {
  [EEventLogType.PingSweep]: faSearch,
  [EEventLogType.PortScanStarted]: faSpinner,
  [EEventLogType.PortScanCompleted]: faFlagCheckered,
  [EEventLogType.DeviceOnline]: faCircleCheck,
  [EEventLogType.DeviceIdle]: faClock,
  [EEventLogType.DeviceOffline]: faTimesCircle,
  [EEventLogType.LocalIPFound]: faGlobe,
  [EEventLogType.LocalNetworkFound]: faNetworkWired,
  [EEventLogType.Warning]: faExclamationCircle,
  [EEventLogType.Alert]: faShieldAlt,
};
