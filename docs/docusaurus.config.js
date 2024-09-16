// docusaurus.config.js

// Import necessary modules
import { themes as prismThemes } from 'prism-react-renderer';

module.exports = {
    title: 'GoTick',
    tagline: 'Flexible, Distributed Scheduling for Go Projects',
    url: 'https://gotick.github.io', // Update this to your domain
    baseUrl: '/core/',
    onBrokenLinks: 'throw',
    onBrokenMarkdownLinks: 'warn',
    favicon: 'img/favicon.ico',
    organizationName: 'misikdmytro', // GitHub org/user name
    projectName: 'gotick', // Repo name
    trailingSlash: false,

    presets: [
        [
            '@docusaurus/preset-classic',
            {
                docs: {
                    sidebarPath: require.resolve('./sidebars.js'), // Link to your sidebar
                    editUrl: 'https://github.com/go-tick/core/edit/main/', // Link to GitHub edit page
                },
                blog: {
                    showReadingTime: true,
                    editUrl:
                        'https://github.com/go-tick/core/edit/main/blog/', // Blog edit link
                },
                theme: {
                    customCss: require.resolve('./src/css/custom.css'),
                },
            },
        ],
    ],

    themeConfig: {
        navbar: {
            title: 'GoTick',
            // logo: {
            //     alt: 'GoTick Logo',
            //     src: 'img/logo.png', // Update this with your logo path
            // },
            items: [
                {
                    type: 'doc',
                    docId: 'intro',
                    position: 'left',
                    label: 'Docs',
                },
                {
                    href: 'https://github.com/go-tick/core',
                    label: 'GitHub',
                    position: 'right',
                },
                {
                    href: 'https://github.com/go-tick/core/issues',
                    label: 'Report an Issue',
                    position: 'right',
                },
            ],
        },
        footer: {
            style: 'dark',
            links: [
                {
                    title: 'Docs',
                    items: [
                        {
                            label: 'Getting Started',
                            to: '/docs/intro',
                        },
                        {
                            label: 'Guides',
                            to: '/docs/guides/creating-your-first-job',
                        },
                    ],
                },
                {
                    title: 'Community',
                    items: [
                        {
                            label: 'GitHub Discussions',
                            href: 'https://github.com/go-tick/core/discussions',
                        },
                    ],
                },
                {
                    title: 'More',
                    items: [
                        {
                            label: 'GitHub',
                            href: 'https://github.com/go-tick/core',
                        },
                        {
                            label: 'Report an Issue',
                            href: 'https://github.com/go-tick/core/issues',
                        },
                    ],
                },
            ],
            copyright: `Copyright © ${new Date().getFullYear()} GoTick, Inc. Built with Docusaurus.`,
        },
        prism: {
            theme: prismThemes.github,
            darkTheme: prismThemes.dracula,
        },
    },
};
