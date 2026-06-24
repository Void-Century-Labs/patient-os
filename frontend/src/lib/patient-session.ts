import type { Patient } from "@/lib/api";

const STORAGE_KEY = "patientos:patient";

export function savePatient(patient: Patient) {
  if (typeof window === "undefined") return;
  window.localStorage.setItem(STORAGE_KEY, JSON.stringify(patient));
}

export function loadPatient(): Patient | null {
  if (typeof window === "undefined") return null;
  const raw = window.localStorage.getItem(STORAGE_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as Patient;
  } catch {
    return null;
  }
}
