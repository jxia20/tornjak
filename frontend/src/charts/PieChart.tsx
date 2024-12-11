import React from "react";
import { PieChart } from "@carbon/charts-react";
import "@carbon/charts/styles.css";
import "@carbon/styles/css/styles.css";
import { connect } from "react-redux";
import { RootState } from "redux/reducers";
import { Alignments, LegendPositions, PieChartOptions } from "@carbon/charts";
import { PieChartEntry } from "components/types";

type PieChartProps = {
  data: PieChartEntry[];
  title: string;
};

type PieChartState = {
  options: PieChartOptions;
};

class PieChart1 extends React.Component<PieChartProps, PieChartState> {
  constructor(props: PieChartProps) {
    super(props);
    this.state = {
      options: this.createChartOptions(props.title),
    };
  }

  // Helper method to create chart options
  createChartOptions = (title: string): PieChartOptions => ({
    title,
    resizable: true,
    height: "300px",
    legend: {
      position: LegendPositions.RIGHT,
      truncation: {
        type: "mid_line",
        threshold: 15,
        numCharacter: 12,
      },
    },
    pie: {
      alignment: Alignments.CENTER,
    },
  });

  componentDidUpdate(prevProps: PieChartProps) {
    if (prevProps.title !== this.props.title) {
      // Update options if title changes
      this.setState({ options: this.createChartOptions(this.props.title) });
    }
  }

  render() {
    const { data } = this.props;
    const { options } = this.state;

    return (
      <div>
        <PieChart data={data} options={options} />
      </div>
    );
  }
}

const mapStateToProps = (state: RootState) => ({});

export default connect(mapStateToProps, null)(PieChart1);
