"use client";

import { useEffect, useRef, useState } from "react";
import { joinCsv, splitCsv } from "@/lib/format";

type Props = {
  id?: string;
  values: string[];
  onChange: (values: string[]) => void;
};

// Keeps the raw typed text as the source of truth so trailing commas/spaces
// aren't stripped mid-typing by re-deriving the input value from the parsed array.
export function CsvInput({ id, values, onChange }: Props) {
  const [text, setText] = useState(() => joinCsv(values));
  const lastEmitted = useRef(text);

  useEffect(() => {
    const incoming = joinCsv(values);
    if (incoming !== lastEmitted.current) {
      setText(incoming);
      lastEmitted.current = incoming;
    }
  }, [values]);

  function handleChange(next: string) {
    setText(next);
    lastEmitted.current = joinCsv(splitCsv(next));
    onChange(splitCsv(next));
  }

  return (
    <input
      id={id}
      value={text}
      onChange={(e) => handleChange(e.target.value)}
    />
  );
}
