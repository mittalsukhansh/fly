import React, { useState, useEffect, useCallback } from 'react';
import HologramGraph from './components/HologramGraph';
import { Layers, Activity, Search } from 'lucide-react';
import './App.css';

function App() {
  const [graphData, setGraphData] = useState({ nodes: [], edges: [] });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [legend, setLegend] = useState({});
  
  // UI Controls
  const [layoutMode, setLayoutMode] = useState('anatomical');
  const [searchMode, setSearchMode] = useState('subgraph');
  const [queryType, setQueryType] = useState('PN');
  const [pathFrom, setPathFrom] = useState('');
  const [pathTo, setPathTo] = useState('');

  const fetchData = async () => {
    if (searchMode === 'subgraph' && !queryType) return;
    if (searchMode === 'path' && (!pathFrom || !pathTo)) return;
    
    setLoading(true);
    setError(null);
    
    try {
      let url = '';
      if (searchMode === 'subgraph') {
        url = `/subgraph?type=${queryType}&layout=${layoutMode}`;
      } else {
        url = `/path?from=${pathFrom}&to=${pathTo}&layout=${layoutMode}&weighted=true`;
      }
      
      console.log('[App] Fetching:', url);
      const res = await fetch(url);
      if (!res.ok) {
        const text = await res.text();
        throw new Error(text || `HTTP ${res.status}`);
      }
      const data = await res.json();
      console.log('[App] Got', (data.nodes || []).length, 'nodes,', (data.edges || []).length, 'edges');
      setGraphData(data);
    } catch (err) {
      console.error('[App] Error:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchData(); }, []);
  useEffect(() => { fetchData(); }, [layoutMode]);

  const handleLegend = useCallback((l) => setLegend(l), []);

  const nodeCount = (graphData.nodes || graphData.Nodes || []).length;
  const edgeCount = (graphData.edges || graphData.Edges || []).length;

  return (
    <div className="app-container">
      {/* 3D Scene */}
      <div className="hologram-layer">
        <HologramGraph graphData={graphData} layoutMode={layoutMode} onLegend={handleLegend} />
      </div>
      
      {/* UI Overlay */}
      <div className="ui-overlay">
        {/* Top Bar */}
        <div className="glass-panel top-panel">
          <div className="logo">
            <Activity className="icon neon-cyan" size={22} />
            <h1>Connectome Hologram</h1>
          </div>
          
          <div className="controls">
            <div className="control-group">
              <Layers className="icon neon-purple" size={16} />
              <select 
                value={layoutMode} 
                onChange={(e) => setLayoutMode(e.target.value)}
                className="glass-select"
              >
                <option value="anatomical">Anatomical Layout</option>
                <option value="logical">Logical Layout</option>
              </select>
            </div>
            
            <div className="control-group">
              <Search className="icon neon-cyan" size={16} />
              <select
                value={searchMode}
                onChange={(e) => setSearchMode(e.target.value)}
                className="glass-select"
              >
                <option value="subgraph">Subgraph</option>
                <option value="path">Path</option>
              </select>
              
              {searchMode === 'subgraph' ? (
                <input 
                  type="text" 
                  className="glass-input"
                  value={queryType} 
                  onChange={(e) => setQueryType(e.target.value)} 
                  placeholder="Neuron Type (e.g. PN)" 
                />
              ) : (
                <>
                  <input 
                    type="text" className="glass-input"
                    value={pathFrom} onChange={(e) => setPathFrom(e.target.value)} 
                    placeholder="From ID" style={{width: '85px'}}
                  />
                  <span className="arrow-icon">→</span>
                  <input 
                    type="text" className="glass-input"
                    value={pathTo} onChange={(e) => setPathTo(e.target.value)} 
                    placeholder="To ID" style={{width: '85px'}}
                  />
                </>
              )}
              
              <button className="glass-btn" onClick={fetchData}>
                {loading ? '...' : 'Extract'}
              </button>
            </div>
          </div>
        </div>

        {/* Status + Legend row */}
        <div className="info-row">
          <div className="status-bar">
            {loading && <span className="loading-text">⟳ Extracting...</span>}
            {error && <span className="error-text">✗ {error}</span>}
            {!loading && !error && nodeCount > 0 && (
              <span className="stats-text">
                ● {nodeCount} neurons · {edgeCount} synapses
              </span>
            )}
          </div>

          {/* Color Legend */}
          {Object.keys(legend).length > 0 && (
            <div className="glass-panel legend-panel">
              {Object.entries(legend).map(([type, info]) => (
                <div key={type} className="legend-item">
                  <span className="legend-dot" style={{ background: info.color }} />
                  <span className="legend-label">{type}</span>
                  <span className="legend-count">{info.count}</span>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Interaction hint */}
      <div className="hint-bar">
        Left-click: rotate · Scroll: zoom · Right-click: pan · Hover: inspect neuron
      </div>
    </div>
  );
}

export default App;
