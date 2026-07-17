import { Resume } from "@/lib/types";

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
    skills: [],
    experience: [],
    projects: [],
    education: [],
    openSourceContributions: [],
    certifications: []
  };
}
