import React from 'react';
import clsx from 'clsx';
import styles from './styles.module.css';

const FeatureList = [
    {
        title: '🕒 Customizable Timeouts',
        description:
            'Assign unique timeouts per job to ensure precise control over execution times.',
    },
    {
        title: '🌐 Distributed Scheduling',
        description:
            'Easily manage jobs across clusters using a centralized database for state tracking.',
    },
    {
        title: '🔌 Extensible Architecture',
        description:
            'Extend GoTick with custom database drivers and tailor it to your project’s needs.',
    },
];

function Feature({ title, description }) {
    return (
        <div className={clsx(styles.feature)}>
            <h3 className={styles.featureTitle}>{title}</h3>
            <p className={styles.featureDescription}>{description}</p>
        </div>
    );
}

export default function HomepageFeatures() {
    return (
        <section className={styles.featuresSection}>
            <div className={styles.featuresContainer}>
                {FeatureList.map((props, idx) => (
                    <Feature key={idx} {...props} />
                ))}
            </div>
        </section>
    );
}
