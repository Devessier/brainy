module.exports = {
  title: "brainy",
  description: "brainy online documentation",

  themeConfig: {
    sidebar: {
      '/': [
        {
          text: 'Introduction',
          children: [
            {
              text: 'Tour of brainy',
              link: '/tour-of-brainy'
            },
          ]
        }
      ]
    }
  }
};
