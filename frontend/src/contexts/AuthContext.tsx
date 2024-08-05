import React, { createContext, useState, useContext, ReactNode } from 'react';

interface AuthContextType {
  isUserLoggedIn: boolean;
  setIsUserLoggedIn: (isLoggedIn: boolean) => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [isUserLoggedIn, setIsUserLoggedIn] = useState<boolean>(false);

  return (
    <AuthContext.Provider value={{ isUserLoggedIn, setIsUserLoggedIn }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
