// Entry for the nginx-served page loaded inside Console iframe (/proxy/<ext>/index.html).
// Not used by ks-console yarn dev; that flow uses src/index.js + registerExtension.
import * as React from 'react';
import { createRoot } from 'react-dom/client';
import App from './App';

const el = document.getElementById('root');
if (el) {
  createRoot(el).render(<App />);
}
