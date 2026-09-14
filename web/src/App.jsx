import React, { useState, useEffect } from 'react';
import HologramGraph from './components/HologramGraph';
import { Layers, Activity, Search } from 'lucide-react';
import './App.css';

function App() {
  const [graphData, setGraphData] = useState({ nodes: [], edges: [] });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  
  // UI Controls
  const [layoutMode, setLayoutMode] = useState('anatomical');
  const [queryType, setQueryType] = useState('PN');
  const [hasQueried, setHasQueried] = useState(false);

  // Background hologram graph (entire connectome very faint)
  const [bgGraphData, setBgGraphData] = useState({ nodes: [], edges: [] });

  const fetchSubgraph = async () => {
    if (!queryType) return;
    setLoading(true);
    setError(null);
    setHasQueried(true);
    
    try {
      const res = await fetch(`/subgraph?type=${queryType}&layout=${layoutMode}`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      setGraphData(data);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    // Fetch initial data on mount
    fetchSubgraph();
  }, []);

  useEffect(() => {
    // If we've already run a query and we just flip the layout, we re-fetch the layout positions
    if (hasQueried) {
      fetchSubgraph();
    }
  }, [layoutMode]);

  return (
    <div className="app-container">
      {/* 3D Hologram Layer */}
      <div className="hologram-layer">
        <HologramGraph graphData={graphData} layoutMode={layoutMode} />
      </div>
      
      {/* Glassmorphism UI Overlay */}
      <div className="ui-overlay">
        <div className="glass-panel top-panel">
          <div className="logo">
            <Activity className="icon neon-cyan" />
            <h1>Connectome Hologram</h1>
          </div>
          
          <div className="controls">
            <div className="control-group">
              <Layers className="icon neon-purple" size={18} />
              <select 
                value={layoutMode} 
                onChange={(e) => setLayoutMode(e.target.value)}
                className="glass-select"
              >
                <option value="anatomical">Anatomical Layout</option>
                <option value="logical">Logical Node2Vec Layout</option>
              </select>
            </div>
            
            <div className="control-group">
              <Search className="icon neon-cyan" size={18} />
              <input 
                type="text" 
                className="glass-input"
                value={queryType} 
                onChange={(e) => setQueryType(e.target.value)} 
                placeholder="Neuron Type (e.g. PN)" 
              />
              <button className="glass-btn" onClick={fetchSubgraph}>
                Extract
              </button>
            </div>
          </div>
        </div>

        {/* Status indicator */}
        {(loading || error) && (
          <div className="status-panel glass-panel">
            {loading && <p className="loading-text">Extracting Subgraph...</p>}
            {error && <p className="error-text">Error: {error}</p>}
          </div>
        )}
      </div>
    </div>
  );
}

export default App;
