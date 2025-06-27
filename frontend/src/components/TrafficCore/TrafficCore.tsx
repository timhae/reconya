import React, { useState, useEffect, useRef, useMemo } from 'react';
import { Device } from '../../models/device.model';
import './TrafficCore.css';

interface TrafficCoreProps {
  devices: Device[];
  localDevice?: Device;
}

interface DataArtery {
  id: string;
  angle: number;
  intensity: number;
  device: Device;
  isActive: boolean;
}

const TrafficCore: React.FC<TrafficCoreProps> = ({ devices, localDevice }) => {
  const [pulseIntensity, setPulseIntensity] = useState(0.5);
  const animationRef = useRef<number>();
  const [isInitialized, setIsInitialized] = useState(false);

  // Helper functions to normalize property access
  const getDeviceIPv4 = (device: Device) => device.ipv4 || device.IPv4 || '';
  const getDeviceStatus = (device: Device) => device.status || device.Status;
  const getDevicePorts = (device: Device) => device.ports || device.Ports || [];
  const getDeviceWebServices = (device: Device) => device.web_services || device.WebServices || [];

  // Memoize network saturation calculation to prevent recalculation on every render
  const networkSaturation = useMemo(() => {
    if (devices.length === 0) return 0.1;

    const onlineDevices = devices.filter(d => getDeviceStatus(d) === 'online').length;
    const totalOpenPorts = devices.reduce((sum, device) => {
      return sum + getDevicePorts(device).filter(port => port.state === 'open').length;
    }, 0);
    const totalWebServices = devices.reduce((sum, device) => {
      return sum + getDeviceWebServices(device).length;
    }, 0);

    // Calculate saturation based on activity metrics
    const deviceRatio = onlineDevices / devices.length;
    const portActivity = Math.min(totalOpenPorts / 50, 1); // Normalize to max 50 ports
    const webActivity = Math.min(totalWebServices / 20, 1); // Normalize to max 20 services

    return Math.min((deviceRatio * 0.4 + portActivity * 0.3 + webActivity * 0.3), 1);
  }, [devices.length, devices]);

  // Memoize data arteries to prevent recreation on every render
  const dataArteries = useMemo(() => {
    if (devices.length === 0) return [];
    
    return devices.map((device, index) => {
      const angle = (index / devices.length) * 360;
      const openPorts = getDevicePorts(device).filter(port => port.state === 'open').length;
      const webServices = getDeviceWebServices(device).length;
      const isOnline = getDeviceStatus(device) === 'online';
      
      // Calculate intensity based on device activity
      const intensity = Math.min((openPorts * 0.1 + webServices * 0.2 + (isOnline ? 0.3 : 0)), 1);

      return {
        id: getDeviceIPv4(device),
        angle,
        intensity,
        device,
        isActive: isOnline && (openPorts > 0 || webServices > 0)
      };
    });
  }, [devices]);

  // Initialize when we first get devices
  useEffect(() => {
    if (devices.length > 0 && !isInitialized) {
      setIsInitialized(true);
    }
  }, [devices.length, isInitialized]);

  // Animation loop for pulsing effect
  useEffect(() => {
    if (!isInitialized || devices.length === 0) return;
    
    let lastTime = 0;
    const targetFPS = 60;
    const frameInterval = 1000 / targetFPS;

    const animate = (currentTime: number) => {
      if (currentTime - lastTime >= frameInterval) {
        const time = currentTime * 0.001; // Slower time progression for smoother animation
        
        // More subtle pulse using cosine for smoother transitions
        const basePulse = (Math.cos(time) + 1) * 0.25 + 0.5; // Range: 0.5 to 1
        const smoothedIntensity = basePulse * networkSaturation;
        
        setPulseIntensity(smoothedIntensity);
        lastTime = currentTime;
      }

      animationRef.current = requestAnimationFrame(animate);
    };

    // Small delay to prevent initial flicker
    const timeoutId = setTimeout(() => {
      animationRef.current = requestAnimationFrame(animate);
    }, 100);

    return () => {
      if (animationRef.current) {
        cancelAnimationFrame(animationRef.current);
      }
      clearTimeout(timeoutId);
    };
  }, [isInitialized, networkSaturation]);

  const getSaturationColor = () => {
    if (networkSaturation < 0.3) return '#00ff00'; // Green - low activity
    if (networkSaturation < 0.7) return '#ffff00'; // Yellow - medium activity
    return '#ff4444'; // Red - high activity
  };

  const getSaturationLevel = () => {
    if (networkSaturation < 0.3) return 'LOW';
    if (networkSaturation < 0.7) return 'MEDIUM';
    return 'HIGH';
  };

  // Don't render complex elements until we have initial data
  if (!isInitialized && devices.length === 0) {
    return (
      <div className="traffic-core-container">
        <div className="traffic-core-header mb-3">
          <h6 className="text-success mb-1">[ NETWORK TRAFFIC CORE ]</h6>
          <div className="d-flex justify-content-between align-items-center">
            <span className="saturation-level text-muted">
              --
            </span>
            <span className="device-count text-muted">
              0/0 DEVICES ACTIVE
            </span>
          </div>
        </div>
        <div className="d-flex justify-content-center" style={{ height: "300px", alignItems: "center" }}>
          <span className="text-muted small">Waiting for data...</span>
        </div>
      </div>
    );
  }

  return (
    <div className="traffic-core-container">
      <div className="traffic-core-header mb-3">
        <div className="d-flex align-items-center mb-1">
          <h6 className="text-success mb-1">[ NETWORK TRAFFIC CORE ]</h6>
        </div>
        <div className="d-flex justify-content-between align-items-center">
          <span className="saturation-level" style={{ color: getSaturationColor() }}>
            SATURATION: {getSaturationLevel()}
          </span>
          <span className="device-count text-muted">
            {devices.filter(d => getDeviceStatus(d) === 'online').length}/{devices.length} DEVICES ACTIVE
          </span>
        </div>
      </div>

      <div className="traffic-core-visualization">
        <svg width="700" height="250" viewBox="0 0 800 300">
          {/* Background grid and gradients */}
          <defs>
            <pattern id="grid" width="20" height="20" patternUnits="userSpaceOnUse">
              <path d="M 20 0 L 0 0 0 20" fill="none" stroke="#00ff0020" strokeWidth="1"/>
            </pattern>
            
            {/* Glow filters */}
            <filter id="glow">
              <feGaussianBlur stdDeviation="3" result="coloredBlur"/>
              <feMerge> 
                <feMergeNode in="coloredBlur"/>
                <feMergeNode in="SourceGraphic"/>
              </feMerge>
            </filter>
            
            <filter id="arteryglow">
              <feGaussianBlur stdDeviation="2" result="coloredBlur"/>
              <feMerge> 
                <feMergeNode in="coloredBlur"/>
                <feMergeNode in="SourceGraphic"/>
              </feMerge>
            </filter>
            
            {/* Globe flat gradient */}
            <radialGradient id="globeGradient" cx="0.5" cy="0.5" r="0.5">
              <stop offset="0%" stopColor={getSaturationColor()} stopOpacity="0.6"/>
              <stop offset="100%" stopColor={getSaturationColor()} stopOpacity="0.3"/>
            </radialGradient>
            
            {/* Network circuit pattern */}
            <pattern id="globeTexture" x="0" y="0" width="15" height="15" patternUnits="userSpaceOnUse">
              <path d="M0,7.5 L15,7.5 M7.5,0 L7.5,15" stroke={getSaturationColor()} strokeWidth="0.2" opacity="0.3"/>
              <circle cx="7.5" cy="7.5" r="1" fill="none" stroke={getSaturationColor()} strokeWidth="0.3" opacity="0.4"/>
              <rect x="6" y="6" width="3" height="3" fill={getSaturationColor()} opacity="0.1"/>
              <circle cx="3" cy="3" r="0.5" fill={getSaturationColor()} opacity="0.6"/>
              <circle cx="12" cy="12" r="0.5" fill={getSaturationColor()} opacity="0.6"/>
            </pattern>
            
            {/* Digital mesh pattern */}
            <pattern id="digitalMesh" x="0" y="0" width="8" height="8" patternUnits="userSpaceOnUse">
              <path d="M0,0 L8,8 M8,0 L0,8" stroke={getSaturationColor()} strokeWidth="0.15" opacity="0.2"/>
              <path d="M4,0 L4,8 M0,4 L8,4" stroke={getSaturationColor()} strokeWidth="0.1" opacity="0.15"/>
            </pattern>
            
          </defs>

          {/* Grid background */}
          <rect width="100%" height="100%" fill="url(#grid)" />

          {/* Data arteries radiating from center */}
          {dataArteries.map((artery) => {
            const x1 = 400;
            const y1 = 150;
            const length = 120 + (artery.intensity * 60);
            const x2 = x1 + Math.cos((artery.angle - 90) * Math.PI / 180) * length;
            const y2 = y1 + Math.sin((artery.angle - 90) * Math.PI / 180) * length;
            
            return (
              <g key={artery.id}>
                {/* Main artery line */}
                <line
                  x1={x1}
                  y1={y1}
                  x2={x2}
                  y2={y2}
                  stroke={artery.isActive ? getSaturationColor() : '#00ff0040'}
                  strokeWidth={2 + (artery.intensity * 3)}
                  filter="url(#arteryglow)"
                  opacity={0.6 + (artery.intensity * 0.4)}
                  className={artery.isActive ? 'data-artery active' : 'data-artery'}
                />
                
                {/* Data flow particles */}
                {artery.isActive && (
                  <>
                    <circle
                      cx={x1 + (x2 - x1) * 0.3}
                      cy={y1 + (y2 - y1) * 0.3}
                      r="2"
                      fill={getSaturationColor()}
                      opacity="0.8"
                      className="data-particle"
                      style={{
                        animation: `flow-${artery.id} 2s linear infinite`
                      }}
                    />
                    <circle
                      cx={x1 + (x2 - x1) * 0.7}
                      cy={y1 + (y2 - y1) * 0.7}
                      r="1.5"
                      fill={getSaturationColor()}
                      opacity="0.6"
                      className="data-particle"
                      style={{
                        animation: `flow-${artery.id} 2.5s linear infinite 0.5s`
                      }}
                    />
                  </>
                )}
                
                {/* Device endpoint */}
                <circle
                  cx={x2}
                  cy={y2}
                  r={3 + (artery.intensity * 2)}
                  fill={artery.isActive ? getSaturationColor() : '#00ff0060'}
                  stroke="#000"
                  strokeWidth="1"
                  opacity={0.8}
                />
                
                {/* Device IP label */}
                <text
                  x={x2 + (x2 > 200 ? 8 : -8)}
                  y={y2 + 4}
                  fontSize="8"
                  fill="#00ff00"
                  textAnchor={x2 > 200 ? 'start' : 'end'}
                  opacity="0.7"
                >
                  {getDeviceIPv4(artery.device).split('.').slice(-1)[0]}
                </text>
              </g>
            );
          })}

          {/* Hexagonal Network Hub */}
          <g className="network-hub">
            {/* Outer hexagonal frame */}
            <g className="hex-frame">
              <polygon
                points="400,90 450,115 450,150 450,185 400,210 350,185 350,150 350,115"
                fill="none"
                stroke={getSaturationColor()}
                strokeWidth="2"
                opacity="0.6"
                strokeDasharray="8,4"
              />
              
              {/* Middle hexagon */}
              <polygon
                points="400,105 435,125 435,150 435,175 400,195 365,175 365,150 365,125"
                fill="none"
                stroke={getSaturationColor()}
                strokeWidth="1.5"
                opacity="0.5"
                strokeDasharray="4,2"
              />
              
              {/* Inner hexagon */}
              <polygon
                points="400,120 420,135 420,150 420,165 400,180 380,165 380,150 380,135"
                fill="none"
                stroke={getSaturationColor()}
                strokeWidth="1"
                opacity="0.4"
                strokeDasharray="2,1"
              />
            </g>
            
            {/* Data connection nodes */}
            <g className="connection-nodes">
              {[0, 60, 120, 180, 240, 300].map((angle, index) => {
                const x = 400 + Math.cos((angle - 90) * Math.PI / 180) * 60;
                const y = 150 + Math.sin((angle - 90) * Math.PI / 180) * 60;
                const isActive = index < dataArteries.length && dataArteries[index]?.isActive;
                
                return (
                  <g key={`node-${index}`}>
                    {/* Connection line to center */}
                    <line
                      x1="400"
                      y1="150"
                      x2={x}
                      y2={y}
                      stroke={isActive ? getSaturationColor() : '#00ff0040'}
                      strokeWidth={isActive ? "2" : "1"}
                      opacity={isActive ? "0.8" : "0.3"}
                      strokeDasharray="3,3"
                    />
                    
                    {/* Node circle */}
                    <circle
                      cx={x}
                      cy={y}
                      r={isActive ? "6" : "4"}
                      fill={isActive ? getSaturationColor() : '#00ff0060'}
                      opacity={isActive ? "0.8" : "0.5"}
                    />
                    
                    
                  </g>
                );
              })}
            </g>
            
            {/* Central processing core */}
            <g className="processing-core">
              {/* Core hexagon */}
              <polygon
                points="400,135 415,142.5 415,157.5 400,165 385,157.5 385,142.5"
                fill={getSaturationColor()}
                opacity="0.3"
                stroke={getSaturationColor()}
                strokeWidth="2"
              />
              
              {/* Inner core circle */}
              <circle
                cx="400"
                cy="150"
                r="8"
                fill={getSaturationColor()}
                opacity="0.8"
                filter="url(#glow)"
              />
              
              {/* Core ring */}
              <circle
                cx="400"
                cy="150"
                r="12"
                fill="none"
                stroke={getSaturationColor()}
                strokeWidth="1"
                opacity="0.6"
                strokeDasharray="2,2"
              />
            </g>
            
            {/* Data flow indicators */}
            <g className="data-flow">
              <circle
                cx="400"
                cy="150"
                r="25"
                fill="none"
                stroke={getSaturationColor()}
                strokeWidth="1"
                opacity="0.3"
                strokeDasharray="2,6"
              />
              
              <circle
                cx="400"
                cy="150"
                r="35"
                fill="none"
                stroke={getSaturationColor()}
                strokeWidth="1"
                opacity="0.2"
                strokeDasharray="4,8"
              />
            </g>
            
            {/* Status display */}
            <g className="status-display">
              <text x="580" y="100" fontSize="8" fill={getSaturationColor()} opacity="0.8">HUB ACTIVE</text>
              <text x="580" y="115" fontSize="7" fill={getSaturationColor()} opacity="0.7">
                {devices.filter(d => getDeviceStatus(d) === 'online').length} NODES
              </text>
              <text x="580" y="130" fontSize="6" fill={getSaturationColor()} opacity="0.6">
                NETWORK STATUS
              </text>
              
              {/* Status indicator */}
              <rect x="565" y="95" width="6" height="6" fill="none" stroke={getSaturationColor()} strokeWidth="1" opacity="0.7"/>
              
              {/* Status bars */}
              <rect x="580" y="140" width="15" height="2" fill={getSaturationColor()} opacity="0.6"/>
              <rect x="580" y="145" width="20" height="2" fill={getSaturationColor()} opacity="0.4"/>
            </g>
            
            {/* Network label */}
            <text
              x="400"
              y="250"
              fontSize="10"
              fill={getSaturationColor()}
              textAnchor="middle"
              opacity="0.9"
            >
              NETWORK HUB
            </text>
          </g>

          
          {/* Circuit board elements */}
          <g className="circuit-elements" opacity="0.4">
            {/* Circuit traces */}
            <path d="M50,100 L80,100 L80,130 L120,130" fill="none" stroke={getSaturationColor()} strokeWidth="1" opacity="0.3" strokeDasharray="3,2"/>
            <path d="M350,100 L320,100 L320,130 L280,130" fill="none" stroke={getSaturationColor()} strokeWidth="1" opacity="0.3" strokeDasharray="3,2"/>
            <path d="M100,350 L100,320 L130,320 L130,280" fill="none" stroke={getSaturationColor()} strokeWidth="1" opacity="0.3" strokeDasharray="3,2"/>
            <path d="M300,350 L300,320 L270,320 L270,280" fill="none" stroke={getSaturationColor()} strokeWidth="1" opacity="0.3" strokeDasharray="3,2"/>
            
            {/* Circuit nodes */}
            <circle cx="80" cy="130" r="2" fill={getSaturationColor()} opacity="0.6">
              <animate attributeName="opacity" values="0.3;0.8;0.3" dur="3s" repeatCount="indefinite"/>
            </circle>
            <circle cx="320" cy="130" r="2" fill={getSaturationColor()} opacity="0.6">
              <animate attributeName="opacity" values="0.3;0.8;0.3" dur="3.2s" repeatCount="indefinite"/>
            </circle>
            <circle cx="130" cy="280" r="2" fill={getSaturationColor()} opacity="0.6">
              <animate attributeName="opacity" values="0.3;0.8;0.3" dur="2.8s" repeatCount="indefinite"/>
            </circle>
            <circle cx="270" cy="280" r="2" fill={getSaturationColor()} opacity="0.6">
              <animate attributeName="opacity" values="0.3;0.8;0.3" dur="3.5s" repeatCount="indefinite"/>
            </circle>
            
            {/* Micro-circuits */}
            <rect x="75" y="125" width="10" height="10" fill="none" stroke={getSaturationColor()} strokeWidth="0.5" opacity="0.4"/>
            <rect x="315" y="125" width="10" height="10" fill="none" stroke={getSaturationColor()} strokeWidth="0.5" opacity="0.4"/>
            <rect x="125" y="275" width="10" height="10" fill="none" stroke={getSaturationColor()} strokeWidth="0.5" opacity="0.4"/>
            <rect x="265" y="275" width="10" height="10" fill="none" stroke={getSaturationColor()} strokeWidth="0.5" opacity="0.4"/>
            
            {/* Data indicators */}
            <circle cx="78" cy="128" r="1" fill={getSaturationColor()} opacity="0.8">
              <animate attributeName="r" values="1;1.5;1" dur="2s" repeatCount="indefinite"/>
            </circle>
            <circle cx="322" cy="132" r="1" fill={getSaturationColor()} opacity="0.8">
              <animate attributeName="r" values="1;1.5;1" dur="2.3s" repeatCount="indefinite"/>
            </circle>
          </g>
          
          {/* Corner brackets for sci-fi frame */}
          <g className="corner-brackets" stroke={getSaturationColor()} fill="none" strokeWidth="1" opacity="0.5">
            {/* Top-left */}
            <path d="M20,20 L40,20 M20,20 L20,40"/>
            {/* Top-right */}
            <path d="M760,20 L780,20 M780,20 L780,40"/>
            {/* Bottom-left */}
            <path d="M20,280 L20,260 M20,280 L40,280"/>
            {/* Bottom-right */}
            <path d="M780,280 L760,280 M780,280 L780,260"/>
          </g>
        </svg>
      </div>

      {/* Network statistics */}
      <div className="traffic-stats mt-2">
        <div className="row text-center">
          <div className="col-4">
            <div className="stat-value text-success">
              {devices.filter(d => getDeviceStatus(d) === 'online').length}
            </div>
            <div className="stat-label text-muted small">ONLINE</div>
          </div>
          <div className="col-4">
            <div className="stat-value" style={{ color: getSaturationColor() }}>
              {Math.round(networkSaturation * 100)}%
            </div>
            <div className="stat-label text-muted small">USAGE</div>
          </div>
          <div className="col-4">
            <div className="stat-value text-warning">
              {devices.reduce((sum, device) => sum + getDevicePorts(device).filter(port => port.state === 'open').length, 0)}
            </div>
            <div className="stat-label text-muted small">PORTS</div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default TrafficCore;