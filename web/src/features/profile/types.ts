export type Profile = {
  id: number;
  username: string;
  email: string | null;
  name: string;
  date_of_birth: string;
  gender: string;
  weight_kg: number | null;
  height_cm: number | null;
  onboarding_complete: boolean;
};

export type UpdateProfilePayload = {
  name?: string;
  date_of_birth?: string;
  gender?: string;
  weight_kg?: number | null;
  height_cm?: number | null;
  onboarding_complete?: boolean;
};

export const GENDER_OPTIONS = [
  { value: "male", label: "Male" },
  { value: "female", label: "Female" },
  { value: "other", label: "Other" },
] as const;
