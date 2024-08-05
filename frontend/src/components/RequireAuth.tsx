import React from 'react';
import { Navigate } from 'react-router-dom';
import useAuth from '../hooks/useAuth';

const RequireAuth: React.FC<{ children: JSX.Element }> = ({ children }) => {
  const isAuthenticated = useAuth();

  if (isAuthenticated === null) {
    return <div>Loading...</div>;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" />;
  }

  return children;
};

export default RequireAuth;
