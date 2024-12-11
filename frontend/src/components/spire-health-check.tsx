import './style.css';
import React, { Component } from 'react';
import { RootState } from 'redux/reducers';
import { connect } from 'react-redux';
import { Tooltip, InlineLoading } from 'carbon-components-react';
import TornjakApi from './tornjak-api-helpers';
import { spireHealthCheckFunc, spireHealthCheckingFunc } from 'redux/actions';

type SpireHealthCheckProp = {
  spireHealthCheckFunc: (globalSpireHealthCheck: boolean) => void;
  globalSpireHealthCheck: boolean;
  spireHealthCheckingFunc: (globalSpireHealthChecking: boolean) => void;
  globalSpireHealthChecking: boolean;
};

type SpireHealthCheckState = {
  timer: NodeJS.Timeout | null;
};

class SpireHealthCheck extends Component<SpireHealthCheckProp, SpireHealthCheckState> {
  TornjakApi: TornjakApi;

  constructor(props: SpireHealthCheckProp) {
    super(props);
    this.TornjakApi = new TornjakApi(props);
    this.state = {
      timer: null,
    };
  }

  componentDidMount() {
    this.startTimer();
    this.checkSpireHealth();
  }

  componentDidUpdate(prevProps: SpireHealthCheckProp) {
    // Perform health check if globalSpireHealthChecking changes
    if (prevProps.globalSpireHealthChecking !== this.props.globalSpireHealthChecking) {
      this.checkSpireHealth();
    }
  }

  componentWillUnmount() {
    // Clear the timer to prevent memory leaks
    if (this.state.timer) {
      clearTimeout(this.state.timer);
    }
  }

  startTimer = () => {
    const timer = setTimeout(() => {
      this.checkSpireHealth();
      this.startTimer(); // Restart timer
    }, 60000); // Default refresh rate: 1 minute
    this.setState({ timer });
  };

  checkSpireHealth = () => {
    const { spireHealthCheckFunc, spireHealthCheckingFunc } = this.props;
    this.TornjakApi.spireHealthCheck(spireHealthCheckFunc, spireHealthCheckingFunc);
  };

  render() {
    const { globalSpireHealthCheck, globalSpireHealthChecking } = this.props;

    const spireStatus = (
      <div>
        {globalSpireHealthChecking ? (
          <InlineLoading description="Checking SPIRE health..." />
        ) : globalSpireHealthCheck ? (
          <p>SPIRE is healthy</p>
        ) : (
          <p>SPIRE is unhealthy</p>
        )}
      </div>
    );

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
          {spireStatus}
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

