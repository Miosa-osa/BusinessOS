// Desktop Theme Store — Theming, appearance settings, color schemes, desktop settings store
import { writable } from "svelte/store";
import { browser } from "$app/environment";
import type {
  BackgroundFit,
  AnimatedBackgroundSettings,
  BootScreenSettings,
  CursorSettings,
} from "./desktopBackgroundStore";
import type { IconStyle, IconLibrary } from "./desktopIconStore";
import type { WindowAnimationSettings } from "./desktopAnimationStore";

interface DesktopSettings {
  backgroundId: string;
  customBackgroundUrl: string | null;
  backgroundFit: BackgroundFit;
  showNoise: boolean;
  iconStyle: IconStyle;
  iconLibrary: IconLibrary;
  iconSize: number; // 32-128, default 64
  showIconLabels: boolean;
  gridSnap: boolean;
  companyName: string; // Dynamic company name for loading screen
  // Icon customization
  iconSpacing: number; // 8-32, default 16
  iconShadow: boolean;
  iconBorder: boolean;
  iconHoverEffect: boolean;
  // New customization settings
  animatedBackground: AnimatedBackgroundSettings;
  bootScreen: BootScreenSettings;
  cursor: CursorSettings;
  windowAnimations: WindowAnimationSettings;
  // Experimental features
  enable3DDesktop: boolean;
}

const defaultSettings: DesktopSettings = {
  backgroundId: "classic-gray",
  customBackgroundUrl: null,
  backgroundFit: "cover",
  showNoise: true,
  iconStyle: "default",
  iconLibrary: "lucide",
  iconSize: 64,
  showIconLabels: true,
  gridSnap: true,
  companyName: "BUSINESS",
  // Icon customization defaults
  iconSpacing: 16,
  iconShadow: true,
  iconBorder: false,
  iconHoverEffect: true,
  // New customization defaults
  animatedBackground: {
    effect: "none",
    intensity: "subtle",
    colors: ["#667eea", "#764ba2"],
    speed: 1,
  },
  bootScreen: {
    logo: { type: "default", color: "#333333" },
    animation: "terminal",
    messages: { enabled: true, custom: [] },
    colors: { background: "#FAFAFA", text: "#333333", accent: "#2563eb" },
    duration: 3,
  },
  cursor: {
    packId: "system",
  },
  windowAnimations: {
    openAnimation: "scale",
    closeAnimation: "fade",
    minimizeAnimation: "scale",
    speed: "normal",
  },
  // Experimental features
  enable3DDesktop: false,
};

function createDesktopStore() {
  // Load from localStorage if available, merging with defaults for any missing fields
  const stored = browser ? localStorage.getItem("desktop-settings") : null;
  let initial: DesktopSettings = defaultSettings;
  if (stored) {
    try {
      const parsed = JSON.parse(stored);
      // Merge with defaults to handle any missing fields from older versions
      initial = { ...defaultSettings, ...parsed };
    } catch (e) {
      console.warn("Failed to parse desktop settings, using defaults");
    }
  }

  const { subscribe, set, update } = writable<DesktopSettings>(initial);

  return {
    subscribe,

    setBackground: (backgroundId: string) => {
      update((state) => {
        const newState = { ...state, backgroundId, customBackgroundUrl: null };
        if (browser) {
          localStorage.setItem("desktop-settings", JSON.stringify(newState));
        }
        return newState;
      });
    },

    setCustomBackground: (url: string) => {
      update((state) => {
        const newState = {
          ...state,
          backgroundId: "custom",
          customBackgroundUrl: url,
        };
        if (browser) {
          localStorage.setItem("desktop-settings", JSON.stringify(newState));
        }
        return newState;
      });
    },

    setBackgroundFit: (fit: BackgroundFit) => {
      update((state) => {
        const newState = { ...state, backgroundFit: fit };
        if (browser) {
          localStorage.setItem("desktop-settings", JSON.stringify(newState));
        }
        return newState;
      });
    },

    toggleNoise: () => {
      update((state) => {
        const newState = { ...state, showNoise: !state.showNoise };
        if (browser) {
          localStorage.setItem("desktop-settings", JSON.stringify(newState));
        }
        return newState;
      });
    },

    setIconStyle: (iconStyle: IconStyle) => {
      update((state) => {
        const newState = { ...state, iconStyle };
        if (browser) {
          localStorage.setItem("desktop-settings", JSON.stringify(newState));
        }
        return newState;
      });
    },

    setIconLibrary: (iconLibrary: IconLibrary) => {
      update((state) => {
        const newState = { ...state, iconLibrary };
        if (browser) {
          localStorage.setItem("desktop-settings", JSON.stringify(newState));
        }
        return newState;
      });
    },

    setIconSize: (iconSize: number) => {
      // Clamp between 32 and 128
      const clampedSize = Math.max(32, Math.min(128, iconSize));
      update((state) => {
        const newState = { ...state, iconSize: clampedSize };
        if (browser) {
          localStorage.setItem("desktop-settings", JSON.stringify(newState));
        }
        return newState;
      });
    },

    toggleIconLabels: () => {
      update((state) => {
        const newState = { ...state, showIconLabels: !state.showIconLabels };
        if (browser) {
          localStorage.setItem("desktop-settings", JSON.stringify(newState));
        }
        return newState;
      });
    },

    toggleGridSnap: () => {
      update((state) => {
        const newState = { ...state, gridSnap: !state.gridSnap };
        if (browser) {
          localStorage.setItem("desktop-settings", JSON.stringify(newState));
        }
        return newState;
      });
    },

    setIconSpacing: (iconSpacing: number) => {
      const clampedSpacing = Math.max(8, Math.min(32, iconSpacing));
      update((state) => {
        const newState = { ...state, iconSpacing: clampedSpacing };
        if (browser) {
          localStorage.setItem("desktop-settings", JSON.stringify(newState));
        }
        return newState;
      });
    },

    setIconShadow: (iconShadow: boolean) => {
      update((state) => {
        const newState = { ...state, iconShadow };
        if (browser) {
          localStorage.setItem("desktop-settings", JSON.stringify(newState));
        }
        return newState;
      });
    },

    setIconBorder: (iconBorder: boolean) => {
      update((state) => {
        const newState = { ...state, iconBorder };
        if (browser) {
          localStorage.setItem("desktop-settings", JSON.stringify(newState));
        }
        return newState;
      });
    },

    setIconHoverEffect: (iconHoverEffect: boolean) => {
      update((state) => {
        const newState = { ...state, iconHoverEffect };
        if (browser) {
          localStorage.setItem("desktop-settings", JSON.stringify(newState));
        }
        return newState;
      });
    },

    setCompanyName: (companyName: string) => {
      update((state) => {
        const newState = { ...state, companyName: companyName.toUpperCase() };
        if (browser) {
          localStorage.setItem("desktop-settings", JSON.stringify(newState));
        }
        return newState;
      });
    },

    // Animated background settings
    setAnimatedBackground: (settings: Partial<AnimatedBackgroundSettings>) => {
      update((state) => {
        const newState = {
          ...state,
          animatedBackground: { ...state.animatedBackground, ...settings },
        };
        if (browser) {
          localStorage.setItem("desktop-settings", JSON.stringify(newState));
        }
        return newState;
      });
    },

    // Boot screen settings
    setBootScreen: (settings: Partial<BootScreenSettings>) => {
      update((state) => {
        const newState = {
          ...state,
          bootScreen: {
            ...state.bootScreen,
            ...settings,
            // Deep merge nested objects
            logo: settings.logo
              ? { ...state.bootScreen.logo, ...settings.logo }
              : state.bootScreen.logo,
            messages: settings.messages
              ? { ...state.bootScreen.messages, ...settings.messages }
              : state.bootScreen.messages,
            colors: settings.colors
              ? { ...state.bootScreen.colors, ...settings.colors }
              : state.bootScreen.colors,
          },
        };
        if (browser) {
          localStorage.setItem("desktop-settings", JSON.stringify(newState));
        }
        return newState;
      });
    },

    // Cursor settings
    setCursor: (settings: Partial<CursorSettings>) => {
      update((state) => {
        const newState = {
          ...state,
          cursor: { ...state.cursor, ...settings },
        };
        if (browser) {
          localStorage.setItem("desktop-settings", JSON.stringify(newState));
        }
        return newState;
      });
    },

    // Window animation settings
    setWindowAnimations: (settings: Partial<WindowAnimationSettings>) => {
      update((state) => {
        const newState = {
          ...state,
          windowAnimations: { ...state.windowAnimations, ...settings },
        };
        if (browser) {
          localStorage.setItem("desktop-settings", JSON.stringify(newState));
        }
        return newState;
      });
    },

    // Experimental features
    toggle3DDesktop: () => {
      update((state) => {
        const newState = { ...state, enable3DDesktop: !state.enable3DDesktop };
        if (browser) {
          localStorage.setItem("desktop-settings", JSON.stringify(newState));
        }
        return newState;
      });
    },

    set3DDesktop: (enabled: boolean) => {
      update((state) => {
        const newState = { ...state, enable3DDesktop: enabled };
        if (browser) {
          localStorage.setItem("desktop-settings", JSON.stringify(newState));
        }
        return newState;
      });
    },

    reset: () => {
      set(defaultSettings);
      if (browser) {
        localStorage.setItem(
          "desktop-settings",
          JSON.stringify(defaultSettings),
        );
      }
    },
  };
}

export const desktopSettings = createDesktopStore();
