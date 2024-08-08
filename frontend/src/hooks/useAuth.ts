// src/hooks/useAuth.ts
import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

const useAuth = () => {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    const checkAuth = async () => {
      const token = localStorage.getItem('token');
      if (!token) {
        setIsAuthenticated(false);
        navigate('/login');
        return;
      }

      const response = await fetch('http://localhost:3008/check-auth', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.status === 401 || response.status === 400) {
        setIsAuthenticated(false);
        navigate('/login');
      } else {
        setIsAuthenticated(true);
      }
    };
    checkAuth();
  }, [navigate]);

  return isAuthenticated;
};

export default useAuth;
