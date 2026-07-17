export function formatDateTime(input?: string): string {
  if (!input) {
    return "-";
  }

  const parsed = new Date(input);
  if (Number.isNaN(parsed.getTime())) {
    return input;
  }

  return parsed.toLocaleString();
}

export function splitCsv(value: string): string[] {
  return value
    .split(",")
    .map((part) => part.trim())
    .filter(Boolean);
}

export function splitMultiline(value: string): string[] {
  return value
    .split("\n")
    .map((part) => part.trim())
    .filter(Boolean);
}

export function joinCsv(values: string[]): string {
  return values.join(", ");
}

export function joinMultiline(values: string[]): string {
  return values.join("\n");
}
