"use client";

import { useMemo } from "react";
import { useSearchParams } from "next/navigation";
import { ResumeEditor } from "@/components/editor/resume-editor";

export function EditorPageClient() {
  const searchParams = useSearchParams();

  const resumeId = useMemo(() => {
    const idParam = searchParams.get("id");
    if (!idParam || idParam === "new") {
      return null;
    }

    const id = Number(idParam);
    if (Number.isNaN(id) || id <= 0) {
      return null;
    }

    return id;
  }, [searchParams]);

  return <ResumeEditor resumeId={resumeId} />;
}
