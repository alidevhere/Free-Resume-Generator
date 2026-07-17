import { ApiEnvelope, Resume, ResumeListItem, ResumeRecord, ResumeSavePayload } from "@/lib/types";

const baseUrl =
  process.env.NEXT_PUBLIC_API_BASE_URL?.replace(/\/$/, "") || "";

async function parseEnvelope<T>(response: Response): Promise<T> {
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}: ${await response.text()}`);
  }

  const payload = (await response.json()) as ApiEnvelope<T>;
  if (!payload.success) {
    throw new Error(payload.message || "Request failed");
  }

  return payload.data;
}

export async function listResumes(): Promise<ResumeListItem[]> {
  const response = await fetch(`${baseUrl}/api/resumes`, {
    cache: "no-store"
  });
  return parseEnvelope<ResumeListItem[]>(response);
}

export async function getResume(id: number): Promise<ResumeRecord> {
  const response = await fetch(`${baseUrl}/api/resumes/${id}`, {
    cache: "no-store"
  });
  return parseEnvelope<ResumeRecord>(response);
}

export async function createResume(name = "Untitled Resume"): Promise<ResumeRecord> {
  const response = await fetch(`${baseUrl}/api/resumes`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name })
  });
  return parseEnvelope<ResumeRecord>(response);
}

export async function saveResume(id: number | null, payload: ResumeSavePayload): Promise<ResumeRecord> {
  if (id === null) {
    const response = await fetch(`${baseUrl}/api/resumes`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload)
    });
    return parseEnvelope<ResumeRecord>(response);
  }

  const response = await fetch(`${baseUrl}/api/resumes/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload)
  });
  return parseEnvelope<ResumeRecord>(response);
}

export async function deleteResume(id: number): Promise<void> {
  const response = await fetch(`${baseUrl}/api/resumes/${id}`, {
    method: "DELETE"
  });

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}: ${await response.text()}`);
  }
}

export async function buildPdf(resume: Resume, signal?: AbortSignal): Promise<Blob> {
  const response = await fetch(`${baseUrl}/api/resume/pdf`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ resume }),
    signal
  });

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}: ${await response.text()}`);
  }

  return response.blob();
}

export async function buildLatex(resume: Resume): Promise<Blob> {
  const response = await fetch(`${baseUrl}/api/resume/latex`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ resume })
  });

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}: ${await response.text()}`);
  }

  return response.blob();
}
