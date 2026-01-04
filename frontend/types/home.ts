export type CategoryNode = {
  id: string;
  title: string;
  children?: CategoryNode[];
};

export type FeaturedCategory = {
  category_id: string;
  title: string;
  href: string;
  priority: "primary" | "secondary";
  order: number;
  icon_key?: string;
};

export type SearchPreset = {
  label: string;
  query: string;
  type?: "good" | "service";
  category_id?: string;
};

export type HomeModel = {
  version: "2";
  tenant_id: string;
  locale: string;
  hero: {
    title: string;
    subtitle?: string;
    searchPlaceholder: string;
    showTypeToggle: boolean;
    showCitySelect: boolean;
    defaultType: "all" | "good" | "service";
  };
  featuredCategories: FeaturedCategory[];
  searchPresets?: SearchPreset[];
};
