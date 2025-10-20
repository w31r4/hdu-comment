import { useState, useEffect, useRef } from 'react';
import { Spin } from 'antd';
import { PictureOutlined } from '@ant-design/icons';

interface LazyImageProps {
    src: string;
    alt: string;
    className?: string;
    placeholder?: string;
    errorFallback?: string;
}

const LazyImage = ({
    src,
    alt,
    className = '',
    placeholder = '/placeholder-image.jpg',
    errorFallback = '/error-image.jpg'
}: LazyImageProps) => {
    const [imageSrc, setImageSrc] = useState(placeholder);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(false);
    const imageRef = useRef<HTMLImageElement>(null);
    const [isInView, setIsInView] = useState(false);

    useEffect(() => {
        const observer = new IntersectionObserver(
            (entries) => {
                entries.forEach((entry) => {
                    if (entry.isIntersecting) {
                        setIsInView(true);
                        observer.disconnect();
                    }
                });
            },
            {
                rootMargin: '50px',
                threshold: 0.1,
            }
        );

        if (imageRef.current) {
            observer.observe(imageRef.current);
        }

        return () => {
            if (imageRef.current) {
                observer.unobserve(imageRef.current);
            }
        };
    }, []);

    useEffect(() => {
        if (!isInView) return;

        const img = new window.Image();

        img.onload = () => {
            setImageSrc(src);
            setLoading(false);
            setError(false);
        };

        img.onerror = () => {
            setImageSrc(errorFallback);
            setLoading(false);
            setError(true);
        };

        img.src = src;

        return () => {
            img.onload = null;
            img.onerror = null;
        };
    }, [src, isInView, errorFallback]);

    return (
        <div className={`lazy-image-container ${className}`} ref={imageRef}>
            {loading && (
                <div className="lazy-image-loading">
                    <Spin size="small" />
                </div>
            )}

            <img
                src={imageSrc}
                alt={alt}
                className={`lazy-image ${loading ? 'lazy-image-loading' : ''} ${error ? 'lazy-image-error' : ''}`}
                loading="lazy"
            />

            {error && (
                <div className="lazy-image-error-overlay">
                    <PictureOutlined style={{ fontSize: 24 }} />
                    <span>图片加载失败</span>
                </div>
            )}
        </div>
    );
};

export default LazyImage;