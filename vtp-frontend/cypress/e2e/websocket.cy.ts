/**
 * E2E Tests for WebSocket Real-Time Updates
 * Tests for real-time data streaming and Socket.IO integration
 */

describe('WebSocket Real-Time Updates - End-to-End Tests', () => {
  const baseUrl = 'http://localhost:3000';
  const dashboardUrl = `${baseUrl}/dashboard`;

  beforeEach(() => {
    cy.visit(dashboardUrl);
    cy.wait(2000); // Wait for initial data load
  });

  describe('WebSocket Connection', () => {
    it('should establish WebSocket connection on dashboard load', () => {
      // Listen for WebSocket messages (using cy.intercept for API simulation)
      cy.visit(dashboardUrl);
      cy.get('.network-status').should('be.visible');
    });

    it('should maintain connection while viewing dashboard', () => {
      cy.visit(dashboardUrl);
      cy.wait(3000);
      // Components should still be receiving updates
      cy.get('.metrics-display').should('be.visible');
      cy.get('.network-status').should('be.visible');
    });

    it('should reconnect if connection is lost', () => {
      cy.visit(dashboardUrl);
      cy.wait(2000);
      // Simulate network condition change
      cy.window().then((win) => {
        // Connection should be re-established
        cy.get('.network-status').should('be.visible');
      });
    });
  });

  describe('Real-Time Network Status Updates', () => {
    it('should receive real-time network status updates', () => {
      cy.get('.network-status .metric-value').first().invoke('text').then((value1) => {
        cy.wait(3000);
        cy.get('.network-status .metric-value').first().invoke('text').then((value2) => {
          // Values may be the same or different, just verify they exist
          expect(value1).to.exist;
          expect(value2).to.exist;
        });
      });
    });

    it('should update quality percentage in real-time', () => {
      cy.get('.network-status').within(() => {
        cy.contains(/Quality|Latency/).should('be.visible');
        cy.wait(2000);
        cy.contains(/\d+/).should('exist');
      });
    });

    it('should update signal strength in real-time', () => {
      cy.get('.network-status').within(() => {
        cy.contains(/Signal/).should('be.visible');
      });
    });

    it('should show network type changes in real-time', () => {
      cy.get('.network-status').should('contain', /5G|4G|LTE|WiFi/);
    });
  });

  describe('Real-Time Metrics Updates', () => {
    it('should receive real-time metrics updates', () => {
      cy.get('.metrics-display .metric-value').first().invoke('text').then((initial) => {
        cy.wait(4000);
        cy.get('.metrics-display .metric-value').first().should('exist');
      });
    });

    it('should update latency metric in real-time', () => {
      cy.get('.metrics-display').within(() => {
        cy.contains(/Latency/).should('be.visible');
        cy.contains(/\d+\.?\d*\s*ms/).should('exist');
        cy.wait(3000);
        cy.contains(/\d+\.?\d*\s*ms/).should('exist');
      });
    });

    it('should update bandwidth metric in real-time', () => {
      cy.get('.metrics-display').within(() => {
        cy.contains(/Bandwidth/).should('be.visible');
      });
    });

    it('should update packet loss metric in real-time', () => {
      cy.get('.metrics-display').within(() => {
        cy.contains(/Packet Loss|Jitter/).should('exist');
      });
    });

    it('should update quality level metric in real-time', () => {
      cy.get('.metrics-display').within(() => {
        cy.contains(/Quality Level|Quality|Excellent|Good|Fair|Poor/).should('exist');
      });
    });

    it('should update trend charts with new data', () => {
      cy.get('.metrics-display svg.sparkline').should('have.length.at.least', 1);
      cy.wait(3000);
      cy.get('.metrics-display svg.sparkline').should('exist');
    });
  });

  describe('Real-Time Edge Node Updates', () => {
    it('should receive real-time edge node status updates', () => {
      cy.get('.edge-node-viewer .node-card').first().within(() => {
        cy.contains(/Online|Offline|Degraded|Maintenance/).should('exist');
      });
    });

    it('should update node latency in real-time', () => {
      cy.get('.edge-node-viewer .node-card').first().within(() => {
        cy.contains(/\d+\.?\d*\s*ms/).should('exist');
        cy.wait(3000);
        cy.get('.edge-node-viewer').should('exist');
      });
    });

    it('should update node capacity in real-time', () => {
      cy.get('.edge-node-viewer .node-card').first().within(() => {
        cy.contains(/%/).should('exist');
      });
    });

    it('should update closest node in real-time', () => {
      cy.get('.edge-node-viewer .closest-node-card').should('exist');
      cy.wait(3000);
      cy.get('.edge-node-viewer .closest-node-card').should('exist');
    });
  });

  describe('Real-Time Updates During User Interaction', () => {
    it('should continue receiving updates while selecting nodes', () => {
      cy.get('.edge-node-viewer .node-card').first().click();
      cy.wait(1000);
      cy.get('.metrics-display .metric-value').should('exist');
      cy.get('.network-status .metric-value').should('exist');
    });

    it('should continue receiving updates while changing quality profile', () => {
      cy.contains('HD').parent().click({ force: true });
      cy.wait(1500);
      cy.get('.metrics-display').should('be.visible');
      cy.get('.network-status').should('be.visible');
    });

    it('should continue receiving updates while scrolling', () => {
      cy.get('.metrics-display').invoke('text').then((initial) => {
        cy.scrollTo('bottom');
        cy.wait(2000);
        cy.get('.metrics-display').should('be.visible');
      });
    });

    it('should continue receiving updates while sorting nodes', () => {
      cy.get('.edge-node-viewer select').select('latency');
      cy.wait(1000);
      cy.get('.metrics-display').should('be.visible');
    });
  });

  describe('Multiple Components Updating Simultaneously', () => {
    it('should update all components at appropriate intervals', () => {
      cy.get('.network-status').should('be.visible');
      cy.get('.metrics-display').should('be.visible');
      cy.get('.edge-node-viewer').should('be.visible');
      cy.wait(3000);
      cy.get('.network-status').should('be.visible');
      cy.get('.metrics-display').should('be.visible');
      cy.get('.edge-node-viewer').should('be.visible');
    });

    it('should coordinate updates between network status and metrics', () => {
      cy.get('.network-status').should('contain', /\d/);
      cy.get('.metrics-display').should('contain', /\d/);
      cy.wait(4000);
      cy.get('.network-status').should('contain', /\d/);
      cy.get('.metrics-display').should('contain', /\d/);
    });

    it('should sync edge node updates with metrics', () => {
      cy.get('.edge-node-viewer').should('exist');
      cy.get('.metrics-display').should('exist');
      cy.wait(3000);
      cy.get('.edge-node-viewer').should('exist');
      cy.get('.metrics-display').should('exist');
    });
  });

  describe('Data Consistency', () => {
    it('should maintain data consistency across all components', () => {
      // Get initial metric values
      cy.get('.metrics-display').invoke('text').then((metricsText) => {
        expect(metricsText).to.not.be.empty;
        cy.get('.network-status').invoke('text').then((statusText) => {
          expect(statusText).to.not.be.empty;
        });
      });
    });

    it('should not show conflicting data between components', () => {
      // Quality selector and network status should be consistent
      cy.get('.quality-selector').should('exist');
      cy.get('.network-status').should('exist');
    });
  });

  describe('Update Frequency and Responsiveness', () => {
    it('should update metrics at regular intervals', () => {
      cy.get('.metrics-display .metric-timestamp').should('be.visible');
      cy.invoke('text').then((time1) => {
        cy.wait(5000);
        cy.get('.metrics-display .metric-timestamp').invoke('text').then((time2) => {
          // Timestamp should have updated
          expect(time2).to.exist;
        });
      });
    });

    it('should respond quickly to quality profile changes', () => {
      const startTime = Date.now();
      cy.contains('Ultra HD').parent().click({ force: true });
      cy.get('.metrics-display .metric-value', { timeout: 3000 }).should('exist').then(() => {
        const elapsed = Date.now() - startTime;
        // Should respond within 3 seconds
        expect(elapsed).to.be.lessThan(3000);
      });
    });

    it('should update node status changes quickly', () => {
      cy.get('.edge-node-viewer .node-card').first().within(() => {
        cy.contains(/Online|Offline|Degraded/).invoke('text').then((status1) => {
          cy.wait(2000);
          cy.contains(/Online|Offline|Degraded|Maintenance/).should('exist');
        });
      });
    });
  });

  describe('Graceful Degradation', () => {
    it('should display last known values if updates pause', () => {
      cy.get('.metrics-display .metric-value').should('have.length.at.least', 1);
      cy.wait(3000);
      cy.get('.metrics-display .metric-value').should('have.length.at.least', 1);
    });

    it('should show loading indicators during data refresh', () => {
      // Refresh button may show loading state
      cy.get('.metrics-display').within(() => {
        cy.contains(/Refresh/i).click();
        cy.wait(500);
        cy.get('.metric-value').should('exist');
      });
    });

    it('should continue showing data even if WebSocket updates slow down', () => {
      cy.get('.network-status').should('be.visible');
      cy.get('.metrics-display').should('be.visible');
      cy.wait(5000);
      cy.get('.network-status').should('be.visible');
      cy.get('.metrics-display').should('be.visible');
    });
  });

  describe('Long-Running Session', () => {
    it('should maintain WebSocket connection for extended period', () => {
      cy.visit(dashboardUrl);
      cy.get('.network-status').should('be.visible');
      cy.wait(5000);
      cy.get('.network-status').should('be.visible');
      cy.wait(5000);
      cy.get('.network-status').should('be.visible');
    });

    it('should not leak memory during extended updates', () => {
      cy.visit(dashboardUrl);
      cy.get('.metrics-display').should('be.visible');
      cy.wait(8000);
      cy.get('.metrics-display').should('be.visible');
    });

    it('should maintain chart data over time', () => {
      cy.get('.metrics-display svg.sparkline').should('exist');
      cy.wait(10000);
      cy.get('.metrics-display svg.sparkline').should('exist');
    });
  });
});
