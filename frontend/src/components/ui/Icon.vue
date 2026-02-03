<script setup lang="ts">
import { computed, type Component, h } from "vue";
import {
  MagnifyingGlassIcon,
  XMarkIcon,
  CheckIcon,
  ClipboardDocumentIcon,
  ChatBubbleLeftRightIcon,
  BellIcon,
  ShoppingBagIcon,
  TruckIcon,
  ArrowPathIcon,
  ExclamationTriangleIcon,
  InboxIcon,
  PhotoIcon,
  EyeIcon,
  PencilIcon,
  TrashIcon,
  PlusIcon,
  ChevronDownIcon,
  ChevronUpIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  Cog6ToothIcon,
  UserIcon,
  ArrowRightOnRectangleIcon,
  HomeIcon,
  CubeIcon,
  ClipboardDocumentListIcon,
  ChartBarIcon,
  DocumentTextIcon,
  FolderIcon,
  LinkIcon,
  ArrowDownTrayIcon,
  PrinterIcon,
  FunnelIcon,
  AdjustmentsHorizontalIcon,
  CalendarIcon,
  ClockIcon,
  StarIcon,
  HeartIcon,
  ShareIcon,
  BookmarkIcon,
  TagIcon,
  CurrencyDollarIcon,
  BuildingStorefrontIcon,
  MapPinIcon,
  PhoneIcon,
  EnvelopeIcon,
  GlobeAltIcon,
  InformationCircleIcon,
  QuestionMarkCircleIcon,
  ExclamationCircleIcon,
  CheckCircleIcon,
  XCircleIcon,
  MinusIcon,
  Bars3Icon,
  EllipsisVerticalIcon,
  EllipsisHorizontalIcon,
} from "@heroicons/vue/24/outline";

// Custom Spinner SVG component - proper loading indicator (circle with gap)
const SpinnerIcon = {
  name: "SpinnerIcon",
  render() {
    return h(
      "svg",
      {
        xmlns: "http://www.w3.org/2000/svg",
        fill: "none",
        viewBox: "0 0 24 24",
      },
      [
        h("circle", {
          class: "opacity-25",
          cx: "12",
          cy: "12",
          r: "10",
          stroke: "currentColor",
          "stroke-width": "4",
        }),
        h("path", {
          class: "opacity-75",
          fill: "currentColor",
          d: "M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z",
        }),
      ],
    );
  },
};

// Also import some solid variants for states
import {
  CheckCircleIcon as CheckCircleSolid,
  ExclamationCircleIcon as ExclamationCircleSolid,
  XCircleIcon as XCircleSolid,
  StarIcon as StarSolid,
  HeartIcon as HeartSolid,
} from "@heroicons/vue/24/solid";

interface Props {
  name: string;
  size?: "xs" | "sm" | "md" | "lg" | "xl";
  spin?: boolean;
  solid?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  size: "md",
  spin: false,
  solid: false,
});

const iconMap: Record<string, Component> = {
  // Search & Navigation
  search: MagnifyingGlassIcon,
  close: XMarkIcon,
  times: XMarkIcon,
  x: XMarkIcon,
  menu: Bars3Icon,
  home: HomeIcon,

  // Actions
  check: CheckIcon,
  copy: ClipboardDocumentIcon,
  edit: PencilIcon,
  pencil: PencilIcon,
  delete: TrashIcon,
  trash: TrashIcon,
  plus: PlusIcon,
  add: PlusIcon,
  minus: MinusIcon,
  download: ArrowDownTrayIcon,
  print: PrinterIcon,
  share: ShareIcon,
  link: LinkIcon,

  // Arrows & Chevrons
  "chevron-down": ChevronDownIcon,
  "chevron-up": ChevronUpIcon,
  "chevron-left": ChevronLeftIcon,
  "chevron-right": ChevronRightIcon,
  refresh: ArrowPathIcon,
  spinner: SpinnerIcon, // Custom circular loading spinner (distinct from refresh)
  loading: SpinnerIcon,
  logout: ArrowRightOnRectangleIcon,

  // Communication
  chat: ChatBubbleLeftRightIcon,
  comments: ChatBubbleLeftRightIcon,
  bell: BellIcon,
  notification: BellIcon,
  email: EnvelopeIcon,
  envelope: EnvelopeIcon,
  phone: PhoneIcon,

  // E-commerce
  "shopping-bag": ShoppingBagIcon,
  bag: ShoppingBagIcon,
  truck: TruckIcon,
  shipping: TruckIcon,
  store: BuildingStorefrontIcon,
  tag: TagIcon,
  dollar: CurrencyDollarIcon,
  price: CurrencyDollarIcon,

  // Status & Alerts
  warning: ExclamationTriangleIcon,
  "exclamation-triangle": ExclamationTriangleIcon,
  error: XCircleIcon,
  success: CheckCircleIcon,
  info: InformationCircleIcon,
  question: QuestionMarkCircleIcon,
  "exclamation-circle": ExclamationCircleIcon,

  // Content
  inbox: InboxIcon,
  image: PhotoIcon,
  photo: PhotoIcon,
  eye: EyeIcon,
  view: EyeIcon,
  document: DocumentTextIcon,
  file: DocumentTextIcon,
  folder: FolderIcon,
  clipboard: ClipboardDocumentListIcon,
  list: ClipboardDocumentListIcon,

  // User & Settings
  user: UserIcon,
  settings: Cog6ToothIcon,
  cog: Cog6ToothIcon,
  filter: FunnelIcon,
  adjustments: AdjustmentsHorizontalIcon,

  // Products & Analytics
  cube: CubeIcon,
  product: CubeIcon,
  chart: ChartBarIcon,
  analytics: ChartBarIcon,

  // Date & Time
  calendar: CalendarIcon,
  clock: ClockIcon,

  // Favorites & Social
  star: StarIcon,
  heart: HeartIcon,
  bookmark: BookmarkIcon,

  // Location
  location: MapPinIcon,
  "map-pin": MapPinIcon,
  globe: GlobeAltIcon,

  // Misc
  "more-vertical": EllipsisVerticalIcon,
  "more-horizontal": EllipsisHorizontalIcon,
  dots: EllipsisVerticalIcon,
};

const solidIconMap: Record<string, Component> = {
  success: CheckCircleSolid,
  check: CheckCircleSolid,
  error: XCircleSolid,
  warning: ExclamationCircleSolid,
  star: StarSolid,
  heart: HeartSolid,
};

const iconComponent = computed(() => {
  if (props.solid && solidIconMap[props.name]) {
    return solidIconMap[props.name];
  }
  return iconMap[props.name] || InboxIcon;
});

const sizeClasses: Record<string, string> = {
  xs: "w-3 h-3",
  sm: "w-4 h-4",
  md: "w-5 h-5",
  lg: "w-6 h-6",
  xl: "w-8 h-8",
};
</script>

<template>
  <component
    :is="iconComponent"
    :class="[sizeClasses[size], spin && 'animate-spin']"
    aria-hidden="true"
  />
</template>
