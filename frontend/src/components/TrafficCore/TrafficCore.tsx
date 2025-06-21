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
        <svg width="300" height="300" viewBox="0 0 400 400">
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
            
            {/* Globe 3D gradients */}
            <radialGradient id="globeGradient" cx="0.3" cy="0.3" r="0.8">
              <stop offset="0%" stopColor={getSaturationColor()} stopOpacity="0.8"/>
              <stop offset="30%" stopColor={getSaturationColor()} stopOpacity="0.6"/>
              <stop offset="70%" stopColor="#000000" stopOpacity="0.4"/>
              <stop offset="100%" stopColor="#000000" stopOpacity="0.8"/>
            </radialGradient>
            
            {/* Globe highlight */}
            <radialGradient id="globeHighlight" cx="0.25" cy="0.25" r="0.4">
              <stop offset="0%" stopColor="#ffffff" stopOpacity="0.6"/>
              <stop offset="50%" stopColor={getSaturationColor()} stopOpacity="0.3"/>
              <stop offset="100%" stopColor="transparent" stopOpacity="0"/>
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
            
            {/* Globe shadow */}
            <radialGradient id="globeShadow" cx="0.5" cy="0.5" r="0.8">
              <stop offset="0%" stopColor="#000000" stopOpacity="0.6"/>
              <stop offset="70%" stopColor="#000000" stopOpacity="0.3"/>
              <stop offset="100%" stopColor="#000000" stopOpacity="0"/>
            </radialGradient>
          </defs>

          {/* Grid background */}
          <rect width="100%" height="100%" fill="url(#grid)" />

          {/* Data arteries radiating from center */}
          {dataArteries.map((artery) => {
            const x1 = 200;
            const y1 = 200;
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

          {/* Central 3D globe */}
          <g className="central-globe">
            {/* Outer pulse ring */}
            <circle
              cx="200"
              cy="200"
              r={60 + (pulseIntensity * 15)}
              fill="none"
              stroke={getSaturationColor()}
              strokeWidth="1"
              opacity={pulseIntensity * 0.3}
              filter="url(#glow)"
              className="pulse-ring"
            />
            
            {/* Inner pulse ring */}
            <circle
              cx="200"
              cy="200"
              r={45 + (pulseIntensity * 8)}
              fill="none"
              stroke={getSaturationColor()}
              strokeWidth="1"
              opacity={pulseIntensity * 0.2}
              className="pulse-ring-inner"
            />
            
            {/* Globe shadow/base */}
            <ellipse
              cx="200"
              cy="235"
              rx="32"
              ry="8"
              fill="url(#globeShadow)"
              opacity="0.6"
            />
            
            {/* Main globe sphere */}
            <circle
              cx="200"
              cy="200"
              r="35"
              fill="url(#globeGradient)"
              className="globe-sphere"
            />
            
            {/* Globe surface texture overlay */}
            <circle
              cx="200"
              cy="200"
              r="35"
              fill="url(#globeTexture)"
              opacity="0.4"
              className="globe-texture"
            />
            
            {/* Digital mesh overlay */}
            <circle
              cx="200"
              cy="200"
              r="35"
              fill="url(#digitalMesh)"
              opacity="0.2"
              className="globe-mesh"
            />
            
            {/* Network grid lines - horizontal */}
            <g className="globe-lines" opacity="0.6">
              {/* Core network rings */}
              <ellipse cx="200" cy="200" rx="35" ry="8" fill="none" stroke={getSaturationColor()} strokeWidth="1" opacity="0.8" strokeDasharray="2,1"/>
              <ellipse cx="200" cy="190" rx="30" ry="6" fill="none" stroke={getSaturationColor()} strokeWidth="0.8" opacity="0.6" strokeDasharray="3,2"/>
              <ellipse cx="200" cy="180" rx="22" ry="4" fill="none" stroke={getSaturationColor()} strokeWidth="0.6" opacity="0.5" strokeDasharray="1,1"/>
              <ellipse cx="200" cy="172" rx="12" ry="2" fill="none" stroke={getSaturationColor()} strokeWidth="0.4" opacity="0.4"/>
              <ellipse cx="200" cy="210" rx="30" ry="6" fill="none" stroke={getSaturationColor()} strokeWidth="0.8" opacity="0.6" strokeDasharray="3,2"/>
              <ellipse cx="200" cy="220" rx="22" ry="4" fill="none" stroke={getSaturationColor()} strokeWidth="0.6" opacity="0.5" strokeDasharray="1,1"/>
              <ellipse cx="200" cy="228" rx="12" ry="2" fill="none" stroke={getSaturationColor()} strokeWidth="0.4" opacity="0.4"/>
            </g>
            
            {/* Network connection paths */}
            <g className="globe-meridians" opacity="0.5">
              <path d="M 200 165 Q 185 200 200 235" fill="none" stroke={getSaturationColor()} strokeWidth="1" strokeDasharray="4,2"/>
              <path d="M 200 165 Q 215 200 200 235" fill="none" stroke={getSaturationColor()} strokeWidth="1" strokeDasharray="4,2"/>
              <path d="M 200 165 Q 175 200 200 235" fill="none" stroke={getSaturationColor()} strokeWidth="0.8" strokeDasharray="2,3"/>
              <path d="M 200 165 Q 225 200 200 235" fill="none" stroke={getSaturationColor()} strokeWidth="0.8" strokeDasharray="2,3"/>
              <path d="M 165 200 Q 200 185 235 200" fill="none" stroke={getSaturationColor()} strokeWidth="0.8" strokeDasharray="3,1"/>
              <path d="M 165 200 Q 200 215 235 200" fill="none" stroke={getSaturationColor()} strokeWidth="0.8" strokeDasharray="3,1"/>
              
              {/* Additional network paths */}
              <path d="M 180 175 Q 200 190 220 175" fill="none" stroke={getSaturationColor()} strokeWidth="0.6" strokeDasharray="1,2" opacity="0.7"/>
              <path d="M 180 225 Q 200 210 220 225" fill="none" stroke={getSaturationColor()} strokeWidth="0.6" strokeDasharray="1,2" opacity="0.7"/>
            </g>
            
            {/* Globe highlight */}
            <circle
              cx="200"
              cy="200"
              r="35"
              fill="url(#globeHighlight)"
              opacity="0.8"
              className="globe-highlight"
            />
            
            {/* Network nodes and data packets */}
            <g className="globe-data-points">
              {/* Primary network nodes */}
              <g className="network-node">
                <circle cx="185" cy="190" r="2" fill="none" stroke={getSaturationColor()} strokeWidth="1" opacity="0.8">
                  <animate attributeName="r" values="2;3;2" dur="3s" repeatCount="indefinite"/>
                </circle>
                <circle cx="185" cy="190" r="1" fill={getSaturationColor()} opacity="0.9"/>
              </g>
              
              <g className="network-node">
                <circle cx="215" cy="185" r="2" fill="none" stroke={getSaturationColor()} strokeWidth="1" opacity="0.6">
                  <animate attributeName="r" values="2;2.5;2" dur="4s" repeatCount="indefinite"/>
                </circle>
                <circle cx="215" cy="185" r="0.8" fill={getSaturationColor()} opacity="0.7"/>
              </g>
              
              <g className="network-node">
                <circle cx="190" cy="210" r="2" fill="none" stroke={getSaturationColor()} strokeWidth="1" opacity="0.7">
                  <animate attributeName="r" values="2;2.8;2" dur="2.5s" repeatCount="indefinite"/>
                </circle>
                <circle cx="190" cy="210" r="1.2" fill={getSaturationColor()} opacity="0.8"/>
              </g>
              
              <g className="network-node">
                <circle cx="210" cy="205" r="2" fill="none" stroke={getSaturationColor()} strokeWidth="1" opacity="0.5">
                  <animate attributeName="r" values="2;2.5;2" dur="3.5s" repeatCount="indefinite"/>
                </circle>
                <circle cx="210" cy="205" r="0.8" fill={getSaturationColor()} opacity="0.6"/>
              </g>
              
              {/* Additional network connection points */}
              <circle cx="175" cy="200" r="1" fill={getSaturationColor()} opacity="0.4">
                <animate attributeName="opacity" values="0.2;0.6;0.2" dur="5s" repeatCount="indefinite"/>
              </circle>
              <circle cx="225" cy="200" r="1" fill={getSaturationColor()} opacity="0.4">
                <animate attributeName="opacity" values="0.2;0.6;0.2" dur="4.5s" repeatCount="indefinite"/>
              </circle>
              <circle cx="200" cy="175" r="1" fill={getSaturationColor()} opacity="0.3">
                <animate attributeName="opacity" values="0.1;0.5;0.1" dur="6s" repeatCount="indefinite"/>
              </circle>
              <circle cx="200" cy="225" r="1" fill={getSaturationColor()} opacity="0.3">
                <animate attributeName="opacity" values="0.1;0.5;0.1" dur="5.5s" repeatCount="indefinite"/>
              </circle>
            </g>
            
            {/* Central network core */}
            <g className="network-core">
              {/* Core outer ring */}
              <circle
                cx="200"
                cy="200"
                r="6"
                fill="none"
                stroke={getSaturationColor()}
                strokeWidth="1"
                opacity="0.8"
                strokeDasharray="1,1"
              >
                <animate attributeName="r" values="6;8;6" dur="4s" repeatCount="indefinite"/>
                <animateTransform attributeName="transform" type="rotate" values="0 200 200;360 200 200" dur="8s" repeatCount="indefinite"/>
              </circle>
              
              {/* Core middle ring */}
              <circle
                cx="200"
                cy="200"
                r="4"
                fill="none"
                stroke={getSaturationColor()}
                strokeWidth="1.5"
                opacity="0.9"
                strokeDasharray="2,1"
              >
                <animateTransform attributeName="transform" type="rotate" values="360 200 200;0 200 200" dur="6s" repeatCount="indefinite"/>
              </circle>
              
              {/* Core center */}
              <circle
                cx="200"
                cy="200"
                r="2.5"
                fill={getSaturationColor()}
                opacity="0.9"
                filter="url(#glow)"
              >
                <animate attributeName="opacity" values="0.7;1;0.7" dur="3s" repeatCount="indefinite"/>
              </circle>
              
              {/* Core data indicators */}
              <circle cx="200" cy="194" r="0.5" fill={getSaturationColor()} opacity="0.6">
                <animate attributeName="cy" values="194;206;194" dur="2s" repeatCount="indefinite"/>
              </circle>
              <circle cx="194" cy="200" r="0.5" fill={getSaturationColor()} opacity="0.6">
                <animate attributeName="cx" values="194;206;194" dur="2.5s" repeatCount="indefinite"/>
              </circle>
              <circle cx="206" cy="200" r="0.5" fill={getSaturationColor()} opacity="0.6">
                <animate attributeName="cx" values="206;194;206" dur="2.5s" repeatCount="indefinite"/>
              </circle>
              <circle cx="200" cy="206" r="0.5" fill={getSaturationColor()} opacity="0.6">
                <animate attributeName="cy" values="206;194;206" dur="2s" repeatCount="indefinite"/>
              </circle>
            </g>
            
            {/* Network core text */}
            <text
              x="200"
              y="260"
              fontSize="8"
              fill={getSaturationColor()}
              textAnchor="middle"
              opacity="0.8"
              className="core-label"
            >
              NETWORK CORE
            </text>
          </g>

          {/* Scanning sweep effect */}
          <g className="scanning-sweep" opacity="0.3">
            <line
              x1="200"
              y1="200"
              x2="350"
              y2="200"
              stroke="#00ff00"
              strokeWidth="1"
              opacity="0.8"
              className="sweep-line"
              transform="rotate(0 200 200)"
              style={{
                animation: 'sweep 4s linear infinite'
              }}
            />
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
            <path d="M360,20 L380,20 M380,20 L380,40"/>
            {/* Bottom-left - rotated 90deg counter-clockwise */}
            <path d="M20,380 L20,360 M20,380 L40,380"/>
            {/* Bottom-right - rotated 90deg clockwise */}
            <path d="M380,380 L360,380 M380,380 L380,360"/>
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