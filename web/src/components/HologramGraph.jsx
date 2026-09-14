import React, { useRef, useEffect, useState } from 'react';
import ForceGraph3D from 'react-force-graph-3d';
import * as THREE from 'three';
import { UnrealBloomPass } from 'three/examples/jsm/postprocessing/UnrealBloomPass.js';

const HologramGraph = ({ graphData, layoutMode }) => {
  const fgRef = useRef();

  // Color palette for the neon aesthetic
  const colors = ['#00e5ff', '#b000ff', '#ff007f', '#39ff14', '#ffff00', '#ff00ff', '#00ffff', '#ff5555'];
  const typeColorMap = useRef({});
  let colorIdx = 0;

  // Transform graph data based on layout mode
  const [displayData, setDisplayData] = useState({ nodes: [], links: [] });

  useEffect(() => {
    const nodesArray = graphData.nodes || graphData.Nodes;
    if (!nodesArray) return;

    const newNodes = nodesArray.map(n => {
      const nType = n.type || n.Type || 'Unknown';
      if (!typeColorMap.current[nType]) {
        typeColorMap.current[nType] = colors[colorIdx % colors.length];
        colorIdx++;
      }

      let x = n.x ?? n.X ?? 0;
      let y = n.y ?? n.Y ?? 0;
      let z = n.z ?? n.Z ?? 0;
      let id = n.id ?? n.Id ?? n.ID;
      
      // We scale the coordinates to spread them nicely in the 3D space
      const scale = 50;
      return {
        id: id,
        type: nType,
        color: typeColorMap.current[nType],
        degree: 5, // We can compute this dynamically if needed
        fx: x * scale,
        fy: y * scale,
        fz: z * scale
      };
    });

    const edgesArray = graphData.edges || graphData.Edges || [];
    const newLinks = edgesArray.map(e => ({
      source: e.source ?? e.Source,
      target: e.target ?? e.Target,
      weight: e.weight ?? e.Weight ?? 1
    }));

    setDisplayData({ nodes: newNodes, links: newLinks });
  }, [graphData, layoutMode]);

  // Setup the Bloom effect (The Hologram Glow)
  useEffect(() => {
    const fg = fgRef.current;
    if (fg) {
      // Enable post-processing for bloom
      const bloomPass = new UnrealBloomPass(new THREE.Vector2(window.innerWidth, window.innerHeight), 1.5, 1, 0.1);
      bloomPass.strength = 1.5;
      bloomPass.radius = 1;
      bloomPass.threshold = 0.1;
      fg.postProcessingComposer().addPass(bloomPass);

      // Disable default physics so the nodes stay in their exact Layout positions
      fg.d3Force('charge', null).d3Force('link', null).d3Force('center', null);
      
      // Zoom out to fit the whole brain structure nicely
      setTimeout(() => fg.zoomToFit(1000), 1000);
    }
  }, []);

  return (
    <ForceGraph3D
      ref={fgRef}
      graphData={displayData}
      backgroundColor="#050510"
      nodeId="id"
      nodeColor="color"
      nodeRelSize={4}
      nodeResolution={16}
      nodeOpacity={0.9}
      linkColor={() => 'rgba(102, 252, 241, 0.15)'}
      linkWidth={1}
      linkOpacity={0.3}
      enableNodeDrag={false}
    />
  );
};

export default HologramGraph;
