import React from 'react';

interface HeartIconProps {
  online: boolean;
}

const HeartIcon: React.FC<HeartIconProps> = ({ online }) => {
  return (
    <span
      className={`heart-icon${online ? ' online' : ' offline'}`}
      title={online ? 'Device online' : 'Device offline'}
      style={{
        display: 'inline-block',
        width: 16,
        height: 16,
        verticalAlign: 'middle',
      }}
    >
      <svg
        width="16"
        height="16"
        viewBox="0 0 24 24"
        fill={online ? '#2ecc40' : '#555'}
        xmlns="http://www.w3.org/2000/svg"
        style={{ filter: online ? 'drop-shadow(0 0 2px #2ecc40)' : 'none' }}
      >
        <path d="M12 21s-6.716-5.686-9.543-8.514C-1.13 9.813 1.4 5.5 6.09 5.5c2.13 0 3.91 1.72 5.91 4.09C14.09 7.22 15.87 5.5 18 5.5c4.69 0 7.22 4.313 3.543 7.986C18.716 15.314 12 21 12 21z"/>
      </svg>
    </span>
  );
};

export default HeartIcon;
