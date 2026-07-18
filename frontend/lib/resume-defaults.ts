import { Resume } from "@/lib/types";

export const DEFAULT_SECTION_ORDER = [
  "summary",
  "education",
  "experience",
  "projects",
  "openSource",
  "skills",
  "certifications",
];

export function createEmptyResume(): Resume {
  return {
    name: "",
    title: "",
    phone: "",
    email: "",
    website: "",
    linkedin: "",
    github: "",
    template: "enhanced-faang-resume",
    summary: "",
    sectionOrder: [...DEFAULT_SECTION_ORDER],
    skills: [],
    experience: [],
    projects: [],
    education: [],
    openSourceContributions: [],
    certifications: []
  };
}
