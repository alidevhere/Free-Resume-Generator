"use client";

import { useRef, useState } from "react";
import { DEFAULT_SECTION_ORDER } from "@/lib/resume-defaults";

const SECTION_LABELS: Record<string, string> = {
  summary: "Summary",
  education: "Education",
  experience: "Experience",
  projects: "Projects",
  openSource: "Open Source Contributions",
  skills: "Technical Skills",
  certifications: "Certifications",
};

type Props = {
  order: string[];
  onChange: (order: string[]) => void;
};

export function SectionOrderEditor({ order, onChange }: Props) {
  const dragIndex = useRef<number | null>(null);
  const [dragOver, setDragOver] = useState<number | null>(null);

  function move(from: number, to: number) {
    const next = [...order];
    const [item] = next.splice(from, 1);
    next.splice(to, 0, item);
    onChange(next);
  }

  function onDragStart(index: number) {
    dragIndex.current = index;
  }

  function onDragEnter(index: number) {
    setDragOver(index);
  }

  function onDrop(index: number) {
    if (dragIndex.current !== null && dragIndex.current !== index) {
      move(dragIndex.current, index);
    }
    dragIndex.current = null;
    setDragOver(null);
  }

  function onDragEnd() {
    dragIndex.current = null;
    setDragOver(null);
  }

  return (
    <div className="section-order-editor">
      <div className="section-order-hint">
        Drag rows or use arrows to reorder sections in the PDF.
      </div>
      <ul className="section-order-list">
        {order.map((key, i) => (
          <li
            key={key}
            className={`section-order-item${dragOver === i ? " drag-over" : ""}`}
            draggable
            onDragStart={() => onDragStart(i)}
            onDragEnter={() => onDragEnter(i)}
            onDragOver={(e) => e.preventDefault()}
            onDrop={() => onDrop(i)}
            onDragEnd={onDragEnd}
          >
            <span className="drag-handle" title="Drag to reorder">
              ⠿
            </span>
            <span className="section-label">{SECTION_LABELS[key] ?? key}</span>
            <span className="section-order-arrows">
              <button
                type="button"
                disabled={i === 0}
                onClick={() => move(i, i - 1)}
                title="Move up"
              >
                ▲
              </button>
              <button
                type="button"
                disabled={i === order.length - 1}
                onClick={() => move(i, i + 1)}
                title="Move down"
              >
                ▼
              </button>
            </span>
          </li>
        ))}
      </ul>
      <button
        type="button"
        className="reset-order-btn"
        onClick={() => onChange([...DEFAULT_SECTION_ORDER])}
      >
        Reset to default order
      </button>
    </div>
  );
}
