// sidebars.js

module.exports = {
    // Define the main sidebar structure
    tutorialSidebar: [
        {
            type: 'category',
            label: 'Getting Started 🚀',
            collapsible: true,
            collapsed: false,
            items: [
                'intro', // Link to intro.md or intro.mdx
                'installation', // Link to installation guide
                'quick-start', // Link to a quick start guide
            ],
        },
        {
            type: 'category',
            label: 'Guides 📘',
            collapsible: true,
            collapsed: true,
            items: [
                'guides/creating-your-first-job', // Step-by-step guide
                'guides/configuring-timeouts', // Guide on setting timeouts
                'guides/distributed-setup', // Guide on distributed scheduling
                {
                    type: 'category',
                    label: 'Advanced Topics 🔧',
                    collapsible: true,
                    collapsed: true,
                    items: [
                        'guides/extending-gotick', // Guide on extending functionality
                        'guides/custom-drivers', // Creating custom database drivers
                        'guides/best-practices', // Best practices for performance
                    ],
                },
            ],
        },
        {
            type: 'link',
            label: 'GitHub Repository',
            href: 'https://github.com/go-tick/core',
        },
        {
            type: 'link',
            label: 'Report an Issue 🐞',
            href: 'https://github.com/go-tick/core/issues',
        },
    ],
};
