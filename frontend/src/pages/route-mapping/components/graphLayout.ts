import {
  forceSimulation,
  forceLink,
  forceManyBody,
  forceCollide,
  forceX,
  forceY,
  type SimulationNodeDatum,
} from "d3-force";
import type { Edge } from "@xyflow/react";
import type { GraphNode } from "./GraphViewHelpers";

/**
 * D3 force-directed layout for positioning graph nodes.
 * Extracted from GraphViewHelpers per SRP — layout algorithm is independent
 * of data transformation logic.
 */

interface SimNode extends SimulationNodeDatum {
  id: string;
  type: "component" | "route";
  x?: number;
  y?: number;
}

export function runForceLayout(nodes: GraphNode[], edges: Edge[]) {
  const simNodes: SimNode[] = nodes.map((n) => ({
    id: n.id,
    type: n.type as "component" | "route",
    x: n.position.x,
    y: n.position.y,
  }));

  const simLinks = edges.map((e) => ({
    source: e.source,
    target: e.target,
  }));

  const simulation = forceSimulation(simNodes)
    .force(
      "link",
      forceLink(simLinks)
        .id((d) => (d as SimNode).id)
        .distance(200),
    )
    .force("charge", forceManyBody().strength(-300))
    .force(
      "collide",
      forceCollide().radius((d) =>
        (d as SimNode).type === "component" ? 100 : 50,
      ),
    )
    .force(
      "x",
      forceX()
        .x((d) => ((d as SimNode).type === "component" ? -300 : 300))
        .strength(0.5),
    )
    .force("y", forceY().strength(0.1));

  simulation.tick(300);

  return nodes.map((node, i) => {
    const simNode = simNodes[i];
    return {
      ...node,
      position: { x: simNode.x || 0, y: simNode.y || 0 },
    };
  });
}
