import './style.css';
import React, { Component } from 'react';
import { connect } from 'react-redux';
import { Tooltip, InlineLoading } from 'carbon-components-react';
import { RootState } from 'redux/reducers';
import TornjakApi from './tornjak-api-helpers';
import { spireHealthCheckFunc, spireHealthCheckingFunc } from 'redux/actions';

type SpireHealthCheckProps = {
  spireHealthCheckFunc: (globalSpireHealthCheck: boolean) => void;
  globalSpireHealthCheck: boolean;
  spireHealthCheckingFunc: (globalSpireHealthChecking: boolean) => void;
  globalSpireHealthChecking: boolean;
};

type SpireHealthCheckState = {
  timer: NodeJS.Timeout | null;
};

class SpireHealthCheck extends Component<SpireHealthCheckProps, SpireHealthCheckState> {
  private tornjakApi: TornjakApi;

  constructor(props: SpireHealthCheckProps) {
    super(props);
    this.tornjakApi = new TornjakApi(props);
    this.state = {
      timer: null,
    };
  }

  componentDidMount() {
    this.initiateHealthCheck();
  }

  componentDidUpdate(prevProps: SpireHealthCheckProps) {
    // Ensure health check is retriggered only when the health checking status changes
    if (
      prevProps.globalSpireHealthChecking !== this.props.globalSpireHealthChecking
    ) {
      this.checkSpireHealth();
    }
  }

  componentWillUnmount() {
    this.clearTimer();
  }

  initiateHealthCheck = () => {
    this.startTimer();
    this.checkSpireHealth();
  };

  startTimer = () => {
    const timer = setTimeout(() => {
      this.checkSpireHealth();
      this.startTimer(); // Recursively continue health checks at intervals
    }, 60000); // 1-minute interval
    this.setState({ timer });
  };

  clearTimer = () => {
    const { timer } = this.state;
    if (timer) {
      clearTimeout(timer);
      this.setState({ timer: null });
    }
  };

  checkSpireHealth = () => {
    const { spireHealthCheckFunc, spireHealthCheckingFunc } = this.props;
    try {
      this.tornjakApi.spireHealthCheck(spireHealthCheckFunc, spireHealthCheckingFunc);
    } catch (error) {
      console.error("Failed to perform SPIRE health check:", error);
    }
  };

  renderSpireStatus = () => {
    const { globalSpireHealthCheck, globalSpireHealthChecking } = this.props;

    if (globalSpireHealthChecking) {
      return <InlineLoading description="Checking SPIRE health..." />;
    }

    return globalSpireHealthCheck ? (
      <p>SPIRE is healthy</p>
    ) : (
      <p>SPIRE is unhealthy</p>
    );
  };

  render() {
    return (
      <div className="health-check">
        <div className="spire-health-refresh-tooltip">
          <Tooltip>
            <p className="spire-health-helper">SPIRE Health Check Status</p>
          </Tooltip>
        </div>
        <div className="health-status-check-container">
          <div className="health-status-check-title">
            <h6>SPIRE: </h6>
          </div>
          {this.renderSpireStatus()}
        </div>
      </div>
    );
  }
}

const mapStateToProps = (state: RootState) => ({
  globalSpireHealthCheck: state.servers.globalSpireHealthCheck,
  globalSpireHealthChecking: state.servers.globalSpireHealthChecking,
});

export default connect(mapStateToProps, { spireHealthCheckFunc, spireHealthCheckingFunc })(SpireHealthCheck);
export { SpireHealthCheck };

