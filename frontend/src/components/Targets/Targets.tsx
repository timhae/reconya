// components/Targets/Targets.tsx
import React from 'react';
import { Routes, Route } from 'react-router-dom';
import TargetsList from './TargetsList';
import AddTarget from './AddTarget';

const Targets: React.FC = () => {
  return (
    <div>
      <Routes>
        <Route path="/" element={<TargetsList />} />
        <Route path="/add" element={<AddTarget />} />
      </Routes>
    </div>
  );
};

export default Targets;
