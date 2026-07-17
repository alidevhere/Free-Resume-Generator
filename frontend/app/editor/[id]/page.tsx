import { notFound } from "next/navigation";
import { ResumeEditor } from "@/components/editor/resume-editor";

type Props = {
  params: Promise<{ id: string }>;
};

export default async function EditorPage({ params }: Props) {
  const resolved = await params;

  if (resolved.id === "new") {
    return <ResumeEditor resumeId={null} />;
  }

  const id = Number(resolved.id);
  if (Number.isNaN(id) || id <= 0) {
    notFound();
  }

  return <ResumeEditor resumeId={id} />;
}
