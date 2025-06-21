import React, { useRef, useEffect } from 'react';
import './MatrixGlobe.scss';

const GLOBE_RADIUS = 100;
const LINE_COUNT = 32;
const CANVAS_SIZE = 300;

function drawMatrixGlobe(ctx: CanvasRenderingContext2D, time: number) {
  ctx.clearRect(0, 0, CANVAS_SIZE, CANVAS_SIZE);

  // Draw glowing globe
  ctx.save();
  ctx.translate(CANVAS_SIZE / 2, CANVAS_SIZE / 2);
  const pulse = 1 + 0.08 * Math.sin(time / 400);
  const gradient = ctx.createRadialGradient(0, 0, GLOBE_RADIUS * 0.2, 0, 0, GLOBE_RADIUS * pulse);
  gradient.addColorStop(0, 'rgba(0,255,128,0.8)');
  gradient.addColorStop(0.7, 'rgba(0,255,64,0.4)');
  gradient.addColorStop(1, 'rgba(0,0,0,0.1)');
  ctx.beginPath();
  ctx.arc(0, 0, GLOBE_RADIUS * pulse, 0, 2 * Math.PI);
  ctx.fillStyle = gradient;
  ctx.shadowColor = '#00ff80';
  ctx.shadowBlur = 30;
  ctx.fill();
  ctx.restore();

  // Draw radiating lines (data arteries)
  ctx.save();
  ctx.translate(CANVAS_SIZE / 2, CANVAS_SIZE / 2);
  for (let i = 0; i < LINE_COUNT; i++) {
    const angle = (2 * Math.PI * i) / LINE_COUNT + (time / 2000);
    const length = GLOBE_RADIUS + 30 + 10 * Math.sin(time / 300 + i);
    ctx.beginPath();
    ctx.moveTo(
      Math.cos(angle) * (GLOBE_RADIUS * 0.9 * pulse),
      Math.sin(angle) * (GLOBE_RADIUS * 0.9 * pulse)
    );
    ctx.lineTo(Math.cos(angle) * length, Math.sin(angle) * length);
    ctx.strokeStyle = 'rgba(0,255,128,0.7)';
    ctx.lineWidth = 2;
    ctx.shadowColor = '#00ff80';
    ctx.shadowBlur = 8;
    ctx.stroke();
  }
  ctx.restore();

  // Optionally: draw matrix code rain around the globe
  // (for simplicity, just some random green dots)
  ctx.save();
  ctx.translate(CANVAS_SIZE / 2, CANVAS_SIZE / 2);
  for (let i = 0; i < 60; i++) {
    const angle = Math.random() * 2 * Math.PI;
    const r = GLOBE_RADIUS + 20 + Math.random() * 30;
    ctx.beginPath();
    ctx.arc(Math.cos(angle) * r, Math.sin(angle) * r, 1.2, 0, 2 * Math.PI);
    ctx.fillStyle = 'rgba(0,255,64,0.5)';
    ctx.fill();
  }
  ctx.restore();
}

const MatrixGlobe: React.FC = () => {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    let animationFrameId: number;
    const render = (time: number) => {
      const canvas = canvasRef.current;
      if (canvas) {
        const ctx = canvas.getContext('2d');
        if (ctx) drawMatrixGlobe(ctx, time);
      }
      animationFrameId = requestAnimationFrame(render);
    };
    animationFrameId = requestAnimationFrame(render);
    return () => cancelAnimationFrame(animationFrameId);
  }, []);

  return (
    <div className="matrix-globe-container">
      <canvas
        ref={canvasRef}
        width={CANVAS_SIZE}
        height={CANVAS_SIZE}
        className="matrix-globe-canvas"
      />
    </div>
  );
};

export default MatrixGlobe;
