import matplotlib.pyplot as plt
import matplotlib.ticker as mticker  
from matplotlib.ticker import FuncFormatter
import numpy as np
import argparse
import csv

plt.style.use('seaborn-paper')

plt.rcParams["font.family"] = "sans-serif"
plt.rcParams["font.sans-serif"] = "Helvetica"
plt.rcParams["font.size"] = 20
plt.rcParams["axes.titlesize"] = 18
plt.rcParams["axes.labelsize"] = 18
plt.rcParams["legend.fontsize"] = 14
plt.rcParams["xtick.labelsize"] = 14
plt.rcParams["ytick.labelsize"] = 14
plt.rc('axes.formatter', useoffset=False)

parser = argparse.ArgumentParser(description='Plot PIR scheme properties.')
parser.add_argument('-f', '--file', nargs='+', action='store', type=str,
                    help='Name of CSV file')
parser.add_argument('-n', '--name', nargs='+', action='store', type=str,
                    help='Name of PIR scheme')
parser.add_argument('-p', '--plot', action='store', type=str,
                    help='Plot to produce', required=True)
parser.add_argument('-q', '--queries', action='store', type=int,
                    help='Number of queries', default=100)
args = parser.parse_args()

def throughput_with_batching(args):
    data = {}
    fig, ax = plt.subplots()
    ind = 0

    for csvfile, scheme in zip(args.file, args.name):
        print(scheme)
        data[scheme] = {}
        data[scheme]['batch_sz'] = []
        data[scheme]['tput'] = []
        data[scheme]['dev'] = []

        with open(csvfile) as f:
            reader = csv.reader(f)
            next(reader) # skip first line

            for row in reader:
                print(row)
                batch_sz = int(row[0])
                good_tput = float(row[1])
                dev = float(row[2])
                data[scheme]['batch_sz'].append(batch_sz)
                data[scheme]['tput'].append(good_tput)
                data[scheme]['dev'].append(dev)

        plt.scatter(data[scheme]['batch_sz'], data[scheme]['tput'], label="_"+scheme, marker='.')
        line, = plt.plot(data[scheme]['batch_sz'], data[scheme]['tput'], label=scheme, marker='.')
        plt.fill_between(data[scheme]['batch_sz'], 
                         [x-y for (x,y) in zip(data[scheme]['tput'], data[scheme]['dev'])], 
                         [x+y for (x,y) in zip(data[scheme]['tput'], data[scheme]['dev'])], 
                         label="_"+scheme, alpha=0.5, facecolor=line.get_color())
        ind += 1

    plt.xlabel("Num. queries per batch")
    plt.ylabel("Throughput (GiB/s)")
    #plt.title("Expected PIR throughput, with increasing batching sizes")
    plt.xscale('log')
    plt.yscale('log')
    ax.set_xticks([1, 4, 16, 64, 256, 1024])
    ax.set_yticks([10*2**10, 100*2**10, 1000*2**10])
    plt.annotate("SimplePIR", (data['SimplePIR']['batch_sz'][2], 
        data['SimplePIR']['tput'][2]*2.2),
                 size=16, ha='left', rotation=37)
    l = len(data['DoublePIR']['batch_sz']) - 3
    plt.annotate("DoublePIR", (data['DoublePIR']['batch_sz'][l-1], 
        data['DoublePIR']['tput'][l-1]*1.6),
                 size=16, ha='left', rotation=0)
    ax.xaxis.set_major_formatter(FuncFormatter(lambda x,pos: ("%d" % x)))
    ax.yaxis.set_major_formatter(FuncFormatter(lambda x,pos: ("%d" % (x/1024))))
    fig.set_size_inches([4, 3])
    plt.tight_layout()
    plt.savefig('throughput_with_batching.pdf')

def main(args):
    if args.plot == "batch_tput":
        return throughput_with_batching(args)

if __name__ == "__main__":
    main(args)
