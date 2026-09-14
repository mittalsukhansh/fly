import React, { useRef, useEffect, useState, useMemo, useCallback } from 'react';
import * as THREE from 'three';
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js';

const HologramGraph = ({ graphData, layoutMode, onLegend }) => {
  const containerRef = useRef(null);
  const rendererRef = useRef(null);
  const sceneRef = useRef(null);
  const cameraRef = useRef(null);
  const controlsRef = useRef(null);
  const animFrameRef = useRef(null);
  const raycasterRef = useRef(new THREE.Raycaster());
  const mouseRef = useRef(new THREE.Vector2());
  const nodeMeshesRef = useRef([]);
  const [tooltip, setTooltip] = useState(null);

  // Color palette
  const colors = ['#00e5ff', '#b000ff', '#ff007f', '#39ff14', '#ffff00', '#ff00ff', '#00ffff', '#ff5555',
                  '#ff9500', '#7b68ee', '#00ff88', '#ff6b6b', '#4ecdc4', '#f7dc6f', '#bb8fce'];
  const typeColorMap = useRef({});
  const colorIdx = useRef(0);

  // Process graph data
  const processedData = useMemo(() => {
    const nodesArray = graphData?.nodes || graphData?.Nodes || [];
    const edgesArray = graphData?.edges || graphData?.Edges || [];

    if (nodesArray.length === 0) return { nodes: [], edges: [], legend: {} };

    // Count connections per node for sizing
    const degreeMap = {};
    for (const e of edgesArray) {
      const src = e.source ?? e.Source;
      const tgt = e.target ?? e.Target;
      degreeMap[src] = (degreeMap[src] || 0) + 1;
      degreeMap[tgt] = (degreeMap[tgt] || 0) + 1;
    }

    // Parse raw node data
    const rawNodes = nodesArray.map(n => {
      const nType = n.type || n.Type || 'Unknown';
      if (!typeColorMap.current[nType]) {
        typeColorMap.current[nType] = colors[colorIdx.current % colors.length];
        colorIdx.current++;
      }
      const id = n.id ?? n.Id ?? n.ID;
      return {
        id,
        x: n.x ?? n.X ?? 0,
        y: n.y ?? n.Y ?? 0,
        z: n.z ?? n.Z ?? 0,
        type: nType,
        color: typeColorMap.current[nType],
        degree: degreeMap[id] || 0,
      };
    });

    // Compute bounding box
    let minX = Infinity, maxX = -Infinity;
    let minY = Infinity, maxY = -Infinity;
    let minZ = Infinity, maxZ = -Infinity;

    for (const n of rawNodes) {
      if (n.x < minX) minX = n.x;
      if (n.x > maxX) maxX = n.x;
      if (n.y < minY) minY = n.y;
      if (n.y > maxY) maxY = n.y;
      if (n.z < minZ) minZ = n.z;
      if (n.z > maxZ) maxZ = n.z;
    }

    const rangeX = (maxX - minX) || 1;
    const rangeY = (maxY - minY) || 1;
    const rangeZ = (maxZ - minZ) || 1;
    const S = 200;

    const nodes = rawNodes.map(n => ({
      ...n,
      nx: ((n.x - minX) / rangeX) * 2 * S - S,
      ny: ((n.y - minY) / rangeY) * 2 * S - S,
      nz: ((n.z - minZ) / rangeZ) * 2 * S - S,
    }));

    // Build id -> index map
    const idToIdx = {};
    nodes.forEach((n, i) => { idToIdx[n.id] = i; });

    const edges = edgesArray
      .map(e => {
        const srcId = e.source ?? e.Source;
        const tgtId = e.target ?? e.Target;
        const w = e.weight ?? e.Weight ?? 1;
        return { srcIdx: idToIdx[srcId], tgtIdx: idToIdx[tgtId], weight: w };
      })
      .filter(e => e.srcIdx !== undefined && e.tgtIdx !== undefined);

    // Build legend
    const legend = {};
    const typeCounts = {};
    for (const n of nodes) {
      typeCounts[n.type] = (typeCounts[n.type] || 0) + 1;
    }
    for (const [type, count] of Object.entries(typeCounts)) {
      legend[type] = { color: typeColorMap.current[type], count };
    }

    return { nodes, edges, legend };
  }, [graphData]);

  // Push legend up to parent
  useEffect(() => {
    if (onLegend && processedData.legend) {
      onLegend(processedData.legend);
    }
  }, [processedData, onLegend]);

  // Setup Three.js scene once
  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    const scene = new THREE.Scene();
    scene.background = new THREE.Color('#050510');
    sceneRef.current = scene;

    const camera = new THREE.PerspectiveCamera(60, container.clientWidth / container.clientHeight, 1, 10000);
    camera.position.set(0, 150, 500);
    cameraRef.current = camera;

    const renderer = new THREE.WebGLRenderer({ antialias: true });
    renderer.setPixelRatio(window.devicePixelRatio);
    renderer.setSize(container.clientWidth, container.clientHeight);
    container.appendChild(renderer.domElement);
    rendererRef.current = renderer;

    const controls = new OrbitControls(camera, renderer.domElement);
    controls.enableDamping = true;
    controls.dampingFactor = 0.08;
    controls.rotateSpeed = 0.6;
    controls.zoomSpeed = 1.2;
    controls.autoRotate = true;
    controls.autoRotateSpeed = 0.4;
    controlsRef.current = controls;

    scene.add(new THREE.AmbientLight(0xffffff, 0.8));

    // Subtle grid for spatial reference
    const gridHelper = new THREE.GridHelper(500, 20, 0x111133, 0x0a0a22);
    gridHelper.position.y = -210;
    scene.add(gridHelper);

    const animate = () => {
      animFrameRef.current = requestAnimationFrame(animate);
      controls.update();
      renderer.render(scene, camera);
    };
    animate();

    const handleResize = () => {
      const w = container.clientWidth;
      const h = container.clientHeight;
      camera.aspect = w / h;
      camera.updateProjectionMatrix();
      renderer.setSize(w, h);
    };
    window.addEventListener('resize', handleResize);

    // Hover detection
    const handleMouseMove = (event) => {
      const rect = renderer.domElement.getBoundingClientRect();
      mouseRef.current.x = ((event.clientX - rect.left) / rect.width) * 2 - 1;
      mouseRef.current.y = -((event.clientY - rect.top) / rect.height) * 2 + 1;

      raycasterRef.current.setFromCamera(mouseRef.current, camera);
      const intersects = raycasterRef.current.intersectObjects(nodeMeshesRef.current);

      if (intersects.length > 0) {
        const obj = intersects[0].object;
        const data = obj.userData;
        setTooltip({
          x: event.clientX,
          y: event.clientY,
          id: data.id,
          type: data.type,
          degree: data.degree,
        });
        document.body.style.cursor = 'pointer';
      } else {
        setTooltip(null);
        document.body.style.cursor = 'default';
      }
    };
    renderer.domElement.addEventListener('mousemove', handleMouseMove);

    return () => {
      window.removeEventListener('resize', handleResize);
      renderer.domElement.removeEventListener('mousemove', handleMouseMove);
      cancelAnimationFrame(animFrameRef.current);
      controls.dispose();
      renderer.dispose();
      if (container.contains(renderer.domElement)) {
        container.removeChild(renderer.domElement);
      }
    };
  }, []);

  // Update scene when data changes
  useEffect(() => {
    const scene = sceneRef.current;
    const camera = cameraRef.current;
    if (!scene || !camera) return;

    // Clear old objects (keep lights and grid)
    const toRemove = [];
    scene.traverse(child => {
      if (child.isMesh || child.isLine || child.isPoints || child.isLineSegments) {
        if (!(child instanceof THREE.GridHelper)) toRemove.push(child);
      }
    });
    toRemove.forEach(obj => {
      obj.geometry?.dispose();
      if (Array.isArray(obj.material)) {
        obj.material.forEach(m => m.dispose());
      } else {
        obj.material?.dispose();
      }
      scene.remove(obj);
    });
    nodeMeshesRef.current = [];

    const { nodes, edges } = processedData;
    if (nodes.length === 0) return;

    // Max degree for sizing
    const maxDegree = Math.max(1, ...nodes.map(n => n.degree));

    // --- Draw nodes ---
    for (const node of nodes) {
      const sizeFactor = 2 + (node.degree / maxDegree) * 5;
      const color = new THREE.Color(node.color);

      // Core sphere
      const geo = new THREE.SphereGeometry(sizeFactor, 16, 16);
      const mat = new THREE.MeshPhongMaterial({
        color: color,
        emissive: color,
        emissiveIntensity: 0.6,
        transparent: true,
        opacity: 0.9,
        shininess: 80,
      });
      const sphere = new THREE.Mesh(geo, mat);
      sphere.position.set(node.nx, node.ny, node.nz);
      sphere.userData = { id: node.id, type: node.type, degree: node.degree };
      scene.add(sphere);
      nodeMeshesRef.current.push(sphere);

      // Outer glow
      const glowGeo = new THREE.SphereGeometry(sizeFactor * 2, 8, 8);
      const glowMat = new THREE.MeshBasicMaterial({
        color: color,
        transparent: true,
        opacity: 0.08,
      });
      const glow = new THREE.Mesh(glowGeo, glowMat);
      glow.position.set(node.nx, node.ny, node.nz);
      scene.add(glow);
    }

    // --- Draw edges ---
    if (edges.length > 0 && edges.length < 15000) {
      const linePositions = [];
      for (const e of edges) {
        const src = nodes[e.srcIdx];
        const tgt = nodes[e.tgtIdx];
        linePositions.push(src.nx, src.ny, src.nz);
        linePositions.push(tgt.nx, tgt.ny, tgt.nz);
      }

      const lineGeometry = new THREE.BufferGeometry();
      lineGeometry.setAttribute('position', new THREE.Float32BufferAttribute(linePositions, 3));
      const lineMaterial = new THREE.LineBasicMaterial({
        color: 0x66fcf1,
        transparent: true,
        opacity: 0.2,
      });
      const lineSegments = new THREE.LineSegments(lineGeometry, lineMaterial);
      scene.add(lineSegments);
    }

    // Animate camera to fit
    camera.position.set(0, 150, 500);
    camera.lookAt(0, 0, 0);

    console.log(`[HologramGraph] Rendered ${nodes.length} nodes, ${edges.length} edges`);
  }, [processedData]);

  return (
    <>
      <div
        ref={containerRef}
        style={{ width: '100%', height: '100%', position: 'absolute', top: 0, left: 0 }}
      />
      {tooltip && (
        <div style={{
          position: 'fixed',
          left: tooltip.x + 14,
          top: tooltip.y - 10,
          background: 'rgba(10, 12, 20, 0.92)',
          border: '1px solid rgba(102, 252, 241, 0.5)',
          borderRadius: '8px',
          padding: '8px 14px',
          color: '#66fcf1',
          fontFamily: "'Space Grotesk', sans-serif",
          fontSize: '0.82rem',
          pointerEvents: 'none',
          zIndex: 100,
          backdropFilter: 'blur(8px)',
          boxShadow: '0 0 16px rgba(102, 252, 241, 0.15)',
        }}>
          <div style={{ fontWeight: 700, marginBottom: 2 }}>Neuron #{tooltip.id}</div>
          <div style={{ color: '#c5c6c7', fontSize: '0.75rem' }}>
            Type: <span style={{ color: '#ff007f' }}>{tooltip.type}</span>
            {' · '}
            Connections: <span style={{ color: '#39ff14' }}>{tooltip.degree}</span>
          </div>
        </div>
      )}
    </>
  );
};

export default HologramGraph;
