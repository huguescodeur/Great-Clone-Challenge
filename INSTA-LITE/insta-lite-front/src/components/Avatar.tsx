interface Props {
  url?: string;
  name: string;
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  ring?: boolean;
}

const sizes = {
  xs: 'w-6 h-6 text-[9px]',
  sm: 'w-8 h-8 text-xs',
  md: 'w-10 h-10 text-sm',
  lg: 'w-14 h-14 text-lg',
  xl: 'w-20 h-20 text-xl',
};

export default function Avatar({ url, name, size = 'md', ring = false }: Props) {
  const initials = name
    .split(' ')
    .map((w) => w[0])
    .join('')
    .slice(0, 2)
    .toUpperCase();

  const ringClass = ring
    ? 'ring-2 ring-offset-2 ring-[#E1306C]'
    : '';

  if (url) {
    return (
      <img
        src={url}
        alt={name}
        className={`${sizes[size]} ${ringClass} rounded-full object-cover flex-shrink-0`}
      />
    );
  }

  return (
    <div
      className={`${sizes[size]} ${ringClass} rounded-full flex items-center justify-center text-white font-semibold flex-shrink-0`}
      style={{ background: 'linear-gradient(45deg,#f09433,#e6683c,#dc2743,#cc2366,#bc1888)' }}
    >
      {initials}
    </div>
  );
}
