import React from 'react';

interface LoadingSpinnerProps {
  size?: 'sm' | 'md' | 'lg';
  message?: string;
  fullscreen?: boolean;
}

/**
 * A reusable loading spinner component that can be used throughout the application
 */
const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({
  size = 'md',
  message = 'Loading...',
  fullscreen = false
}) => {
  // Calculate spinner size
  const spinnerSize = {
    sm: 'spinner-border-sm',
    md: '',
    lg: 'spinner-border-lg'
  }[size];

  // If fullscreen, render the spinner in the center of the screen
  if (fullscreen) {
    return (
      <div className="position-fixed top-0 start-0 w-100 h-100 d-flex justify-content-center align-items-center" style={{ backgroundColor: 'rgba(0, 0, 0, 0.7)', zIndex: 9999 }}>
        <div className="text-center">
          <div className={`spinner-border text-success ${spinnerSize}`} role="status">
            <span className="visually-hidden">Loading...</span>
          </div>
          {message && <div className="mt-3 text-success">{message}</div>}
        </div>
      </div>
    );
  }

  // Otherwise, render the spinner inline
  return (
    <div className="text-center my-4">
      <div className={`spinner-border text-success ${spinnerSize}`} role="status">
        <span className="visually-hidden">Loading...</span>
      </div>
      {message && <div className="mt-2 text-success">{message}</div>}
    </div>
  );
};

export default LoadingSpinner;