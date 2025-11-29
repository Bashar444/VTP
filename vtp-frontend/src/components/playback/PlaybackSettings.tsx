import { Settings } from 'lucide-react';
import { cn } from '@/utils/cn';

interface PlaybackSettingsProps {
  currentSpeed: number;
  availableSpeeds: number[];
  onSpeedChange: (speed: number) => void;
  onClose?: () => void;
  className?: string;
}

export const PlaybackSettings: React.FC<PlaybackSettingsProps> = ({
  currentSpeed,
  availableSpeeds,
  onSpeedChange,
  onClose,
  className,
}) => {
  return (
    <div className={cn('bg-gray-800 rounded-lg p-4 text-white max-w-xs', className)}>
      <div className="flex items-center justify-between mb-4">
        <h3 className="font-semibold flex items-center gap-2">
          <Settings className="w-4 h-4" />
          Playback Settings
        </h3>
        {onClose && (
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-white transition-colors"
          >
            ✕
          </button>
        )}
      </div>

      <div className="space-y-3">
        <div>
          <label className="block text-sm text-gray-300 mb-2">Playback Speed</label>
          <div className="grid grid-cols-3 gap-2">
            {availableSpeeds.map(speed => (
              <button
                key={speed}
                onClick={() => {
                  onSpeedChange(speed);
                  onClose?.();
                }}
                className={cn(
                  'px-3 py-2 rounded-lg text-sm font-medium transition-colors',
                  currentSpeed === speed
                    ? 'bg-blue-600 text-white'
                    : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
                )}
              >
                {speed}x
              </button>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};

interface QualitySelectorProps {
  qualities: Array<{ height: number; bitrate: number; name: string }>;
  currentQuality: string;
  onQualityChange: (height: number) => void;
  className?: string;
}

export const QualitySelector: React.FC<QualitySelectorProps> = ({
  qualities,
  currentQuality,
  onQualityChange,
  className,
}) => {
  return (
    <div className={cn('bg-gray-800 rounded-lg p-4 text-white max-w-xs', className)}>
      <h3 className="font-semibold mb-3">Quality</h3>
      <div className="space-y-2">
        <button
          onClick={() => onQualityChange(0)}
          className={cn(
            'w-full px-4 py-2 text-left rounded-lg text-sm font-medium transition-colors',
            currentQuality === 'auto'
              ? 'bg-blue-600 text-white'
              : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
          )}
        >
          Auto (Recommended)
        </button>
        {qualities.map(quality => (
          <button
            key={quality.height}
            onClick={() => onQualityChange(quality.height)}
            className={cn(
              'w-full px-4 py-2 text-left rounded-lg text-sm font-medium transition-colors',
              currentQuality === quality.name
                ? 'bg-blue-600 text-white'
                : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
            )}
          >
            <div className="flex items-center justify-between">
              <span>{quality.name}</span>
              <span className="text-xs text-gray-400">
                {(quality.bitrate / 1000000).toFixed(1)} Mbps
              </span>
            </div>
          </button>
        ))}
      </div>
    </div>
  );
};

interface SubtitleSelectorProps {
  subtitles: Array<{ language: string; url: string }>;
  currentSubtitle?: string;
  onSubtitleChange: (language: string) => void;
  className?: string;
}

export const SubtitleSelector: React.FC<SubtitleSelectorProps> = ({
  subtitles,
  currentSubtitle,
  onSubtitleChange,
  className,
}) => {
  return (
    <div className={cn('bg-gray-800 rounded-lg p-4 text-white max-w-xs', className)}>
      <h3 className="font-semibold mb-3">Subtitles</h3>
      <div className="space-y-2">
        <button
          onClick={() => onSubtitleChange('')}
          className={cn(
            'w-full px-4 py-2 text-left rounded-lg text-sm font-medium transition-colors',
            !currentSubtitle
              ? 'bg-blue-600 text-white'
              : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
          )}
        >
          Off
        </button>
        {subtitles.map(subtitle => (
          <button
            key={subtitle.language}
            onClick={() => onSubtitleChange(subtitle.language)}
            className={cn(
              'w-full px-4 py-2 text-left rounded-lg text-sm font-medium transition-colors',
              currentSubtitle === subtitle.language
                ? 'bg-blue-600 text-white'
                : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
            )}
          >
            {subtitle.language.toUpperCase()}
          </button>
        ))}
      </div>
    </div>
  );
};
