import './style.css';
import { RootState } from 'redux/reducers';
import { connect } from 'react-redux';
import { Tooltip, InlineLoading } from 'carbon-components-react';
import TornjakApi from './tornjak-api-helpers';
import { spireHealthCheckFunc, spireHealthCheckingFunc } from 'redux/actions';

type SpireHealthCheckProp = {
  spireHealthCheckFunc: (globalSpireHealthCheck: boolean) => void,
  globalSpireHealthCheck: boolean,
  spireHealthCheckingFunc: (globalSpireHealthCheck: boolean) => void,
  globalSpireHealthChecking: boolean,
};

type SpireHealthCheckState = {
  timer: NodeJS.Timeout | null,
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
    this.TornjakApi.spireHealthCheck(this.props.spireHealthCheckFunc, this.props.spireHealthCheckingFunc);
  }

  componentDidUpdate(prevProps: SpireHealthCheckProp, prevState: SpireHealthCheckState) {
    if (prevState.timer !== this.state.timer) {
      this.TornjakApi.spireHealthCheck(this.props.spireHealthCheckFunc, this.props.spireHealthCheckingFunc);
    }
  }

  startTimer = () => {
    const timer = setTimeout(() => {
      this.setState({ timer: new Date() });
      this.startTimer(); // Restart timer
    }, 60000); // Default refresh rate: 1 minute
    this.setState({ timer });
  };

  render() {
    const { globalSpireHealthCheck, globalSpireHealthChecking } = this.props;
    let spireStatus = (
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

