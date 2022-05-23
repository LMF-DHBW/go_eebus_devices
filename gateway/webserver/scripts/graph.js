var cy = cytoscape({
    container: document.getElementById('cy'), // container to render in
    elements: [
        { group: 'nodes', data: { id: 'noop', resolved: false } },
        {
          group: 'nodes',
          data: { id: 'collect data', resolved: false }
        },
        {
          group: 'nodes',
          data: { id: 'send to waylay', resolved: false }
        },
        { group: 'nodes', data: { id: 'send to BB', resolved: false } },
        { group: 'nodes', data: { id: 'notify me', resolved: false } },
        {
          group: 'edges',
          data: {
            id: 'noop-collect data',
            target: 'collect data',
            source: 'noop',
            directed: 'false'
          }
        },
        {
          group: 'edges',
          data: {
            id: 'collect data-send to waylay',
            target: 'send to waylay',
            source: 'collect data',
            directed: 'false'
          }
        },
        {
          group: 'edges',
          data: {
            id: 'collect data-send to BB',
            target: 'send to BB',
            source: 'collect data',
            directed: 'false'
          }
        },
        {
          group: 'edges',
          data: {
            id: 'send to waylay-notify me',
            target: 'notify me',
            source: 'send to waylay',
            directed: 'false'
          }
        },
        {
          group: 'edges',
          data: {
            id: 'send to BB-notify me',
            target: 'notify me',
            source: 'send to BB',
            directed: 'false'
          }
        }
      ],
    style: [ // the stylesheet for the graph
        {
            selector: 'node',
            style: {
                'background-color': '#666',
                'label': 'data(id)'
            }
        },
        {
            selector: 'edge',
            style: {
                'width': 3,
                'line-color': '#ccc',
                'target-arrow-color': '#ccc',
                'target-arrow-shape': 'triangle',
                'curve-style': 'bezier'
            }
        }
    ],
    layout: {
        name: 'cose',
        rankDir: 'LR',
        nodeDimensionsIncludeLabels: true
    }
});

cy.getElementById('send to BB-notify me').style({'source-arrow-shape': 'none', 'target-arrow-shape': 'none'});
//cy.getElementById('n1n2').style({'target-arrow-shape': 'triangle-backcurve'});
//cy.getElementById('n2n3').style({'source-arrow-shape': 'triangle-backcurve'});

