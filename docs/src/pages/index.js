import React from 'react';
import clsx from 'clsx';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import styles from './index.module.css';
import HomepageFeatures from '../components/HomepageFeatures';

function HomepageHeader() {
    const { siteConfig } = useDocusaurusContext();
    return (
        <header className={clsx('hero', styles.heroBanner)}>
            <div className={styles.headerContainer}>
                <h1 className={styles.heroTitle}>{siteConfig.title}</h1>
                <p className={styles.heroSubtitle}>{siteConfig.tagline}</p>
                <div className={styles.buttons}>
                    <Link
                        className={clsx(
                            'button button--primary button--lg',
                            styles.ctaButton
                        )}
                        to="docs/intro"
                    >
                        🚀 Get Started
                    </Link>
                    <Link
                        className={clsx(
                            'button button--secondary button--lg',
                            styles.githubButton
                        )}
                        to="https://github.com/go-tick/core"
                    >
                        ⭐ Star on GitHub
                    </Link>
                </div>
            </div>
        </header>
    );
}

function Testimonial({ text, author }) {
    return (
        <div className={styles.testimonial}>
            <p className={styles.testimonialText}>"{text}"</p>
            <p className={styles.testimonialAuthor}>- {author}</p>
        </div>
    );
}

function HomepageTestimonials() {
    const testimonials = [
        {
            text: 'GoTick has transformed how we manage scheduled tasks in our Go projects.',
            author: 'Jane Doe, CTO of TechCorp',
        },
        {
            text: 'The flexibility and extensibility of GoTick are unmatched. A must-have for Go developers!',
            author: 'John Smith, Lead Developer at DevSolutions',
        },
    ];

    return (
        <section className={styles.testimonials}>
            <h2>What Our Users Say</h2>
            <div className={styles.testimonialContainer}>
                {testimonials.map((testimonial, index) => (
                    <Testimonial key={index} {...testimonial} />
                ))}
            </div>
        </section>
    );
}

function HomepageCTA() {
    return (
        <section className={styles.ctaSection}>
            <h2>Ready to Supercharge Your Go Projects?</h2>
            <p>
                Try GoTick now and see the difference in your scheduling
                management!
            </p>
            <Link className="button button--primary button--lg" to="docs/intro">
                🚀 Get Started with GoTick
            </Link>
        </section>
    );
}

export default function Home() {
    const { siteConfig } = useDocusaurusContext();
    return (
        <Layout
            title={`Welcome to ${siteConfig.title}`}
            description="GoTick: Flexible, Distributed Scheduling for Go Projects"
        >
            <HomepageHeader />
            <main>
                <HomepageFeatures />
                <HomepageTestimonials />
                <HomepageCTA />
            </main>
        </Layout>
    );
}
