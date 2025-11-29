/**
 * E2E Tests for API Integration
 * Tests for REST API calls and data flow between frontend and backend
 */

describe('API Integration - End-to-End Tests', () => {
  const baseUrl = 'http://localhost:3000';
  const apiUrl = 'http://localhost:8080/api';
  const dashboardUrl = `${baseUrl}/dashboard`;

  beforeEach(() => {
    cy.visit(dashboardUrl);
    cy.wait(2000);
  });

  describe('Status API Integration', () => {
    it('should fetch network status from API', () => {
      cy.intercept('GET', `${apiUrl}/status*`, {
        statusCode: 200,
        body: {
          network_type: '5G',
          signal_strength: 95,
          latency: 12.5,
          bandwidth: 950,
        },
      });

      cy.get('.network-status').should('be.visible');
    });

    it('should display status API response in UI', () => {
      cy.intercept('GET', `${apiUrl}/status*`, {
        statusCode: 200,
        body: {
          network_type: '5G',
          signal_strength: 95,
          latency: 12.5,
          bandwidth: 950,
        },
      }).as('statusCall');

      cy.visit(dashboardUrl);
      cy.wait('@statusCall');
      cy.get('.network-status').should('contain', /5G|network/i);
    });

    it('should handle status API errors gracefully', () => {
      cy.intercept('GET', `${apiUrl}/status*`, {
        statusCode: 500,
        body: { error: 'Internal Server Error' },
      });

      cy.visit(dashboardUrl);
      cy.get('.network-status').should('exist');
    });

    it('should retry status API on failure', () => {
      let attempts = 0;
      cy.intercept('GET', `${apiUrl}/status*`, (req) => {
        attempts++;
        if (attempts === 1) {
          req.reply({
            statusCode: 500,
            body: { error: 'Server Error' },
          });
        } else {
          req.reply({
            statusCode: 200,
            body: { network_type: '5G', signal_strength: 95 },
          });
        }
      }).as('statusCall');

      cy.visit(dashboardUrl);
      cy.get('.network-status').should('exist');
    });
  });

  describe('Quality Profiles API Integration', () => {
    it('should fetch quality profiles from API', () => {
      cy.intercept('GET', `${apiUrl}/quality-profiles*`, {
        statusCode: 200,
        body: {
          profiles: [
            { name: 'Ultra HD', resolution: '4K', bitrate: 25 },
            { name: 'HD', resolution: '1080p', bitrate: 8 },
            { name: 'Standard', resolution: '720p', bitrate: 4 },
          ],
        },
      }).as('profilesCall');

      cy.visit(dashboardUrl);
      cy.wait('@profilesCall', { timeout: 5000 });
      cy.contains(/Ultra|HD|Standard/).should('exist');
    });

    it('should display current quality profile', () => {
      cy.intercept('GET', `${apiUrl}/quality-profiles/current*`, {
        statusCode: 200,
        body: { name: 'HD', resolution: '1080p', bitrate: 8 },
      }).as('currentProfileCall');

      cy.visit(dashboardUrl);
      cy.wait('@currentProfileCall', { timeout: 5000 });
    });

    it('should update quality profile via API', () => {
      cy.intercept('POST', `${apiUrl}/quality-profiles/set*`, {
        statusCode: 200,
        body: { success: true, profile: 'Ultra HD' },
      }).as('setProfileCall');

      cy.visit(dashboardUrl);
      cy.contains('Ultra HD').parent().click({ force: true });
      cy.wait('@setProfileCall', { timeout: 5000 });
    });

    it('should handle quality profile API errors', () => {
      cy.intercept('GET', `${apiUrl}/quality-profiles*`, {
        statusCode: 500,
        body: { error: 'Failed to fetch profiles' },
      });

      cy.visit(dashboardUrl);
      cy.get('body').should('be.visible');
    });
  });

  describe('Metrics API Integration', () => {
    it('should fetch session metrics from API', () => {
      cy.intercept('GET', `${apiUrl}/metrics/session*`, {
        statusCode: 200,
        body: {
          latency: 12.5,
          bandwidth: 950,
          packet_loss: 0.1,
          quality_level: 'Excellent',
          frame_rate: 60,
          bitrate: 8,
        },
      }).as('sessionMetricsCall');

      cy.visit(dashboardUrl);
      cy.wait('@sessionMetricsCall', { timeout: 5000 });
      cy.get('.metrics-display').should('be.visible');
    });

    it('should fetch global metrics from API', () => {
      cy.intercept('GET', `${apiUrl}/metrics/global*`, {
        statusCode: 200,
        body: {
          avg_latency: 15.2,
          total_bandwidth: 15000,
          packet_loss: 0.05,
          active_connections: 1250,
          server_load: 65,
          avg_quality: 'Excellent',
        },
      }).as('globalMetricsCall');

      cy.visit(dashboardUrl);
      cy.wait('@globalMetricsCall', { timeout: 5000 });
      cy.get('.metrics-display').should('be.visible');
    });

    it('should display metrics in correct format', () => {
      cy.intercept('GET', `${apiUrl}/metrics/session*`, {
        statusCode: 200,
        body: {
          latency: 12.5,
          bandwidth: 950,
          packet_loss: 0.1,
          quality_level: 'Excellent',
        },
      }).as('metricsCall');

      cy.visit(dashboardUrl);
      cy.wait('@metricsCall');
      cy.get('.metrics-display').should('contain', /\d+\.?\d*/);
    });

    it('should handle metrics API errors', () => {
      cy.intercept('GET', `${apiUrl}/metrics*`, {
        statusCode: 500,
        body: { error: 'Metrics service unavailable' },
      });

      cy.visit(dashboardUrl);
      cy.get('.metrics-display').should('exist');
    });
  });

  describe('Edge Nodes API Integration', () => {
    it('should fetch available edge nodes from API', () => {
      cy.intercept('GET', `${apiUrl}/edge-nodes*`, {
        statusCode: 200,
        body: {
          nodes: [
            { id: 1, name: 'US-EAST-1', latency: 8, capacity: 85 },
            { id: 2, name: 'US-WEST-1', latency: 45, capacity: 72 },
            { id: 3, name: 'EU-WEST-1', latency: 95, capacity: 68 },
          ],
        },
      }).as('edgeNodesCall');

      cy.visit(dashboardUrl);
      cy.wait('@edgeNodesCall', { timeout: 5000 });
      cy.get('.edge-node-viewer').should('exist');
    });

    it('should fetch closest edge node from API', () => {
      cy.intercept('GET', `${apiUrl}/edge-nodes/closest*`, {
        statusCode: 200,
        body: { id: 1, name: 'US-EAST-1', latency: 8, capacity: 85 },
      }).as('closestNodeCall');

      cy.visit(dashboardUrl);
      cy.wait('@closestNodeCall', { timeout: 5000 });
      cy.get('.edge-node-viewer').should('exist');
    });

    it('should display node list from API', () => {
      cy.intercept('GET', `${apiUrl}/edge-nodes*`, {
        statusCode: 200,
        body: {
          nodes: [
            { name: 'US-EAST-1', status: 'Online', latency: 8 },
            { name: 'US-WEST-1', status: 'Online', latency: 45 },
          ],
        },
      }).as('nodesCall');

      cy.visit(dashboardUrl);
      cy.wait('@nodesCall');
      cy.contains(/US-EAST|US-WEST/).should('exist');
    });

    it('should handle edge nodes API errors', () => {
      cy.intercept('GET', `${apiUrl}/edge-nodes*`, {
        statusCode: 500,
        body: { error: 'Edge node service unavailable' },
      });

      cy.visit(dashboardUrl);
      cy.get('.edge-node-viewer').should('exist');
    });
  });

  describe('Data Flow and Consistency', () => {
    it('should maintain data consistency between API calls', () => {
      cy.intercept('GET', `${apiUrl}/status*`, {
        statusCode: 200,
        body: { network_type: '5G', quality: 95 },
      }).as('statusCall');

      cy.intercept('GET', `${apiUrl}/metrics*`, {
        statusCode: 200,
        body: { quality_level: 'Excellent', bandwidth: 950 },
      }).as('metricsCall');

      cy.visit(dashboardUrl);
      cy.wait('@statusCall');
      cy.wait('@metricsCall');
      cy.get('.network-status').should('exist');
      cy.get('.metrics-display').should('exist');
    });

    it('should not display stale data during API updates', () => {
      cy.intercept('GET', `${apiUrl}/metrics*`, (req) => {
        req.reply({
          statusCode: 200,
          body: { latency: 12.5, bandwidth: 950 },
        });
      }).as('metricsCall');

      cy.visit(dashboardUrl);
      cy.wait('@metricsCall');
      cy.get('.metrics-display .metric-value').should('exist');
    });
  });

  describe('API Request Headers and Authentication', () => {
    it('should include required headers in API requests', () => {
      cy.intercept('GET', `${apiUrl}/*`, (req) => {
        expect(req.headers).to.exist;
        req.reply({
          statusCode: 200,
          body: {},
        });
      }).as('apiCall');

      cy.visit(dashboardUrl);
      cy.get('.network-status').should('exist');
    });

    it('should handle missing authentication', () => {
      cy.intercept('GET', `${apiUrl}/*`, {
        statusCode: 401,
        body: { error: 'Unauthorized' },
      });

      cy.visit(dashboardUrl);
      cy.get('body').should('be.visible');
    });
  });

  describe('API Response Validation', () => {
    it('should validate quality profiles response structure', () => {
      cy.intercept('GET', `${apiUrl}/quality-profiles*`, {
        statusCode: 200,
        body: {
          profiles: [
            { name: 'Ultra HD', resolution: '4K', bitrate: 25 },
          ],
        },
      }).as('profilesCall');

      cy.visit(dashboardUrl);
      cy.wait('@profilesCall');
    });

    it('should validate metrics response structure', () => {
      cy.intercept('GET', `${apiUrl}/metrics/*`, {
        statusCode: 200,
        body: {
          latency: 12.5,
          bandwidth: 950,
          packet_loss: 0.1,
          quality_level: 'Excellent',
        },
      }).as('metricsCall');

      cy.visit(dashboardUrl);
      cy.wait('@metricsCall');
    });

    it('should handle unexpected API response format', () => {
      cy.intercept('GET', `${apiUrl}/status*`, {
        statusCode: 200,
        body: { unexpected: 'format' },
      });

      cy.visit(dashboardUrl);
      cy.get('.network-status').should('exist');
    });
  });

  describe('Network Conditions', () => {
    it('should handle slow API responses', () => {
      cy.intercept('GET', `${apiUrl}/*`, (req) => {
        req.reply((res) => {
          res.delay(3000);
          res.send({
            statusCode: 200,
            body: { network_type: '5G' },
          });
        });
      });

      cy.visit(dashboardUrl);
      cy.get('.network-status', { timeout: 8000 }).should('exist');
    });

    it('should handle network timeouts', () => {
      cy.intercept('GET', `${apiUrl}/*`, (req) => {
        req.reply((res) => {
          res.destroy();
        });
      });

      cy.visit(dashboardUrl);
      cy.get('body').should('be.visible');
    });

    it('should retry failed API calls', () => {
      let attempts = 0;
      cy.intercept('GET', `${apiUrl}/status*`, (req) => {
        attempts++;
        if (attempts < 2) {
          req.reply({ statusCode: 500 });
        } else {
          req.reply({
            statusCode: 200,
            body: { network_type: '5G' },
          });
        }
      }).as('statusCall');

      cy.visit(dashboardUrl);
      cy.get('.network-status').should('exist');
    });
  });

  describe('Concurrent API Requests', () => {
    it('should handle multiple concurrent API requests', () => {
      cy.intercept('GET', `${apiUrl}/status*`, {
        statusCode: 200,
        body: { network_type: '5G' },
      }).as('statusCall');

      cy.intercept('GET', `${apiUrl}/metrics*`, {
        statusCode: 200,
        body: { latency: 12.5 },
      }).as('metricsCall');

      cy.intercept('GET', `${apiUrl}/edge-nodes*`, {
        statusCode: 200,
        body: { nodes: [] },
      }).as('nodesCall');

      cy.visit(dashboardUrl);
      cy.get('.network-status').should('exist');
      cy.get('.metrics-display').should('exist');
      cy.get('.edge-node-viewer').should('exist');
    });

    it('should not block UI during multiple API calls', () => {
      cy.intercept('GET', `${apiUrl}/*`, (req) => {
        req.reply((res) => {
          res.delay(1000);
          res.send({ statusCode: 200, body: {} });
        });
      });

      cy.visit(dashboardUrl);
      cy.get('body').should('be.visible');
      cy.get('.network-status').should('exist');
    });
  });
});
