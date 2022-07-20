import matplotlib.pyplot as plt
import matplotlib.ticker as mticker  
from matplotlib.ticker import StrMethodFormatter
from matplotlib.transforms import ScaledTranslation
import matplotlib.patches as mpatches
from matplotlib.ticker import FuncFormatter
from matplotlib.patches import Ellipse
import numpy as np
import argparse
import csv

import custom_style
import tufte

#print(plt.rcParams.keys())
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

schemes = ['SealPIR', 'FastPIR', 'OnionPIR',
           'Spiral', 'SpiralPack', 'SpiralStream', 'SpiralStreamPack']

offline_comm = {}
online_comm = {}
tput = {}

offline_comm['SealPIR'] = 5000
online_comm['SealPIR'] = 93 + 186
tput['SealPIR'] = 101

offline_comm['FastPIR'] = 66
online_comm['FastPIR'] = 34000 + 66
tput['FastPIR'] = 226

offline_comm['OnionPIR'] = 5000
online_comm['OnionPIR'] = 262 + 131
tput['OnionPIR'] = 63

offline_comm['Spiral'] = 16000
online_comm['Spiral'] = 14 + 21
tput['Spiral'] = 271

#offline_comm['SpiralPack'] = 20000
#online_comm['SpiralPack'] = 14 + 20
#tput['SpiralPack'] = 269

offline_comm['SpiralStream'] = 344
online_comm['SpiralStream'] = 15000 + 20
tput['SpiralStream'] = 496

#offline_comm['SpiralStreamPack'] = 16000
#online_comm['SpiralStreamPack'] = 30000 + 101
#tput['SpiralStreamPack'] = 1450


def xaxis(x, pos=None):
    'The two args are the value and tick position'
    return '$%d' % (x)

def gbs(x, pos=None):
    'The two args are the value and tick position'
    return '%.1f GB/s' % (x)

def plot_tradeoff(args):
    num_queries = args.queries
    data = {}

    plt.rcParams["font.size"] = 18
    plt.rcParams["axes.titlesize"] = 18
    plt.rcParams["axes.labelsize"] = 18
    plt.rcParams["legend.fontsize"] = 18
    plt.rcParams["xtick.labelsize"] = 18
    plt.rcParams["ytick.labelsize"] = 18

    fig, ax = plt.subplots()
    for csvfile, scheme in zip(args.file, args.name):
        data[scheme] = {}
        data[scheme]['item_sz'] = []
        data[scheme]['tput'] = []
        data[scheme]['tput_dev'] = []
        data[scheme]['total_comm'] = []

        with open(csvfile) as f:
            reader = csv.reader(f)
            next(reader) # skip first line

            for row in reader:
                print(row)
                if row[0] == 'N':
                    break

                item_sz = int(row[1])
                comm = float(row[4])/num_queries + float(row[5])
                t = float(row[2]) # in MB/s 
                t_dev = float(row[3])

                if scheme != "SimplePIR" and scheme != "DoublePIR":
                    comm /= 1024.0 # in KB

                data[scheme]['item_sz'].append(item_sz) 
                data[scheme]['total_comm'].append(comm * 1024)
                data[scheme]['tput'].append(t * 1024 * 1024) 
                data[scheme]['tput_dev'].append(t_dev)

        i1 = np.argmin(data[scheme]['total_comm'])
        i2 = np.argmax(data[scheme]['tput'])
        print(scheme, ": best throughput of ", data[scheme]['tput'][i2], 
              " at entry sz ", data[scheme]['item_sz'][i2], " bits")

        pts = plt.scatter([data[scheme]['total_comm'][i] for i in [i1, i2]], 
                          [data[scheme]['tput'][i] for i in [i1, i2]], 
                          label=scheme, marker='x', s=100)
        c = pts.get_facecolors()[0].tolist()

        x_shift = 0
        y_shift = 0

        if (scheme == "DoublePIR"):
            i2 = i1 # Plot other point
            y_shift = -0.4 * data[scheme]['tput'][i2]
            x_shift = data[scheme]['total_comm'][i2]
        if (scheme == "SpiralPack"):
            # Put label in bounds
            x_shift = data[scheme]['total_comm'][i2]*2
            y_shift = 0.15 * data[scheme]['tput'][i2]
        if (scheme == "OnionPIR"):
            x_shift = data[scheme]['total_comm'][i2]*2.5
            y_shift = -0.4 * data[scheme]['tput'][i2]
        if (scheme == "Spiral"):
            x_shift = data[scheme]['total_comm'][i2]*1.2
            y_shift = -0.42 * data[scheme]['tput'][i2]
        if (scheme == "SimplePIR"):
            y_shift = -0.1 * data[scheme]['tput'][i2]
            x_shift = data[scheme]['total_comm'][i2]*4
        if (scheme == "SealPIR"):
            y_shift = 0.1 * data[scheme]['tput'][i2]
            x_shift = data[scheme]['total_comm'][i2]/1.5

        if (scheme == "SimplePIR") or (scheme == "DoublePIR"):
            plt.annotate(scheme,#+"\n(d="+str(data[scheme]['item_sz'][i2])+" bits)",
                        (data[scheme]['total_comm'][i2]*0.97 + x_shift, 
                        data[scheme]['tput'][i2]*1.05 + y_shift), 
                        size=16, ha='left', color=c)# fontweight='bold')
        else:
            plt.annotate(scheme,#+"\n(d="+str(data[scheme]['item_sz'][i2])+" bits)",
                         (data[scheme]['total_comm'][i2]*0.97 + x_shift, 
                         data[scheme]['tput'][i2]*1.05 + y_shift), 
                         size=16, color=c)

    # Add arrow
    x_tail = 10000 * 1000
    y_tail = 0.07 * (10**9)
    x_head = 3000 * (1000)
    y_head = 0.16 * (10**9)
    plt.annotate("Better", (x_tail*1.1, y_tail*1.15), color="blue", 
                 size=16, ha='left', rotation=28, fontweight='bold')


    # Ellipse centre coordinates
    x, y = x_tail*2.6, y_tail*0.6
    # use the axis scale tform to figure out how far to translate
    ell_offset = ScaledTranslation(x, y, ax.transScale)
    # construct the composite tform
    ell_tform = ell_offset + ax.transLimits + ax.transAxes
    # Create the ellipse centred on the origin, apply the composite tform

    ax.add_patch(Ellipse(xy=(0, 0), width=5.15, height=3.1, facecolor="#FFFFE0", fill=True, lw=2, edgecolor="gray", zorder=-25, transform=ell_tform, linestyle=":"))

    prop = dict(arrowstyle="<|-,head_width=0.4,head_length=0.8",
            shrinkA=0,shrinkB=0,facecolor="blue",edgecolor="blue",linewidth=4)
    plt.annotate("", xy=(x_tail,y_tail), xytext=(x_head,y_head), #transform = ax.transAxes, 
                 color="blue", arrowprops=prop)


    # Flip axis
    plt.gca().invert_xaxis()
    for axis in [ax.xaxis, ax.yaxis]:
        ax.yaxis.set_major_formatter(mticker.FuncFormatter(lambda y,pos: ('{{:.{:1d}f}}'.format(int(np.maximum(-np.log10(y),0)))).format(y)))
    #ax.ticklabel_format(style='plain')
    plt.xlabel("Communication\n (amortized over " + str(num_queries) + " queries)")
    plt.ylabel("Throughput (MiB/s)")
    plt.xscale('log')
    plt.yscale('log')

    ax.set_yticks([10**9 * x for x in [0.125, 0.25, 0.5, 1, 2, 4, 8]])
    ax.set_xticks([100*2**10, 2**20, 10*2**20])
    #ax.set_xlim([None, 400*2**20])

    ax.xaxis.set_major_formatter(FuncFormatter(custom_style.megabytes))
    ax.yaxis.set_major_formatter(FuncFormatter(custom_style.millions))
    custom_style.remove_chart_junk(plt, ax)
    plt.tight_layout()
    plt.savefig('tradeoff_'+str(num_queries)+'_4.pdf')


def compare_all(args):
    num_queries = args.queries
    data = {}

    fig, ax = plt.subplots()
    plt.rcParams["font.size"] = 25

    for csvfile, scheme in zip(args.file, args.name):
        data[scheme] = {}
        data[scheme]['item_sz'] = []
        data[scheme]['tput'] = []
        data[scheme]['tput_dev'] = []
        data[scheme]['total_comm'] = []

        print("Reading ", csvfile)
        with open(csvfile) as f:
            reader = csv.reader(f)
            next(reader) # skip first line

            for row in reader:
                print(row)
                if row[0] == 'N':
                    break
                item_sz = int(row[1])
                comm = float(row[4])/num_queries + float(row[5])
                t = float(row[2]) # in MB.s
                t_dev = float(row[3])

                if scheme != "SimplePIR" and scheme != "DoublePIR":
                    comm /= 1024.0

                data[scheme]['item_sz'].append(item_sz)
                data[scheme]['total_comm'].append(comm)
                data[scheme]['tput'].append(t)
                data[scheme]['tput_dev'].append(t_dev)

        plt.scatter(data[scheme]['item_sz'], data[scheme]['total_comm'], label="_"+scheme, marker='.')
        p=plt.plot(data[scheme]['item_sz'], data[scheme]['total_comm'], label=scheme)

        offset = 1.0

        if scheme == "SpiralStream": offset = 0.4
        if scheme == "Spiral": offset = 1
        if scheme == "SpiralPack": offset = 0.5
        if scheme == "SealPIR": offset = 7.3
        if scheme == "OnionPIR": offset = 1.2
        if scheme == "SimplePIR": offset = 0.8

        color = p[-1].get_color()
        if scheme != "FastPIR":
            plt.text(40*1000, offset*data[scheme]['total_comm'][15], "%s" % (scheme), size='13', color=color)
        else:
            plt.text(3, 16777280, "FastPIR", size='13',color=color)
    plt.xlim(np.min(data['SealPIR']['item_sz']), 
             np.max(data['SealPIR']['item_sz'])+10000)
    #plt.legend(loc='center left', bbox_to_anchor=(0, 1.5),
    #            ncol=3, fancybox=True)
    plt.xlabel("Entry size (in bits)")
    plt.ylabel("Communication\n (amortized over "+str(num_queries)+" queries)")
    plt.xscale('log')
    plt.yscale('log')
    #ax.xaxis.set_major_formatter(FuncFormatter(custom_style.megabytes10))
    ax.xaxis.set_major_formatter(FuncFormatter(custom_style.as_number))
    ax.yaxis.set_major_formatter(FuncFormatter(custom_style.megabytes))
    #ax.xaxis.set_major_formatter(FuncFormatter(custom_style.reformat_large_tick_values))
    #ax.set_xticks([(4**i) for i in range(8)])
    ax.set_xticks([(10**i) for i in range(5)])
    ax.set_yticks([((10**(i%3))*(2**(10*int(i/3)))) for i in range(2, 8)])
    print("TICKING AT: ", [((10**(i%3))*(2**(10*int(i/3)))) for i in range(2, 8)])
    custom_style.remove_chart_junk(plt, ax)
    plt.tight_layout()
    plt.savefig('all_communication_'+str(num_queries)+'_4.pdf')

    # Save legend in a separate figure
    fig_leg = plt.figure()
    ax_leg = fig_leg.add_subplot(111)
    ax_leg.legend(*ax.get_legend_handles_labels(), loc="center",
                  ncol=3, fancybox=True)
    ax_leg.axis("off")
    custom_style.remove_chart_junk(plt, ax)
    plt.tight_layout()
    fig_leg.savefig("legend.pdf")

    # Now make the throughput plot
    plt.clf()
    fig, ax = plt.subplots()
    plt.rcParams["font.size"] = 25
    for scheme in args.name:
        plt.scatter(data[scheme]['item_sz'], data[scheme]['tput'], label="_"+scheme, marker='.')
        line, = plt.plot(data[scheme]['item_sz'], data[scheme]['tput'], label=scheme)
        plt.fill_between(data[scheme]['item_sz'],
                         [x-y for (x,y) in zip(data[scheme]['tput'],data[scheme]['tput_dev'])],
                         [x+y for (x,y) in zip(data[scheme]['tput'],data[scheme]['tput_dev'])],
                         facecolor=line.get_color(), alpha=0.5)
        #plt.text(800, 1.08*data[scheme]['total_comm'][11], scheme, size='13')
        color = line.get_color()

        offset = 1.0
        if scheme == "DoublePIR": offset = 0.6
        if scheme == "OnionPIR": offset = 0.7
        if scheme == "Spiral": offset = 0.7
        if scheme == "FastPIR": offset = 0.9
        plt.text(40*1000, offset*data[scheme]['tput'][0], "%s" % (scheme), size='13', color=color)

    plt.xlim(np.min(data['SealPIR']['item_sz']), 
             np.max(data['SealPIR']['item_sz'])+10000)
    #plt.legend(loc='upper center', bbox_to_anchor=(0.5, 1.05),
    #      ncol=3, fancybox=True)
    plt.xlabel("Entry size (in bits)")
    plt.ylabel("Throughput (MiB/s)")
    plt.xscale('log')
    plt.yscale('log')
    ax.set_xticks([1,10,100,1000,10000])
    ax.xaxis.set_major_formatter(FuncFormatter(custom_style.as_number))
    ax.yaxis.set_major_formatter(FuncFormatter(custom_style.as_number))
    plt.tight_layout()
    custom_style.remove_chart_junk(plt, ax)
    plt.savefig('all_tput_'+str(num_queries)+'_4.pdf')

def communication(args):
    num_queries = args.queries
    data = {}

    fig, ax = plt.subplots()

    ind = 0
    for csvfile, scheme in zip(args.file, args.name):
        data[scheme] = {}
        data[scheme]['item_sz'] = []
        data[scheme]['tput'] = []
        data[scheme]['tput_dev'] = []
        #data[scheme]['offline_comm'] = []
        #data[scheme]['online_comm'] = []
        data[scheme]['total_comm'] = []

        with open(csvfile) as f:
            reader = csv.reader(f)
            next(reader) # skip first line

            for row in reader:
                print(row)
                item_sz = int(row[1])
                comm = float(row[4])/num_queries + float(row[5])
                t = float(row[2])
                t_dev = float(row[3])

                data[scheme]['item_sz'].append(item_sz)
                data[scheme]['total_comm'].append(comm)
                data[scheme]['tput'].append(t)
                data[scheme]['tput_dev'].append(t_dev)

        plt.scatter(data[scheme]['item_sz'], data[scheme]['total_comm'], label="_"+scheme, marker='.', c=colors[ind])
        plt.plot(data[scheme]['item_sz'], data[scheme]['total_comm'], label=scheme, c=colors[ind])
        plt.text(800, 1.08*data[scheme]['total_comm'][11], scheme, size='13')
        ind += 1

    for scheme in offline_comm:
        comm = float(offline_comm[scheme])/num_queries + float(online_comm[scheme])
        plt.axhline(y=comm, linestyle='-', c="gray")
        plt.text(1.5, comm*1.05, scheme, size='13')
        ind += 1

    plt.xlim(1, 4096)
    #plt.legend()
    plt.xlabel("Entry size (in bits)")
    plt.ylabel("Per-query, total communication (KB)")
    plt.xscale('log')
    plt.yscale('log')
    plt.tight_layout()
    custom_style.remove_chart_junk(plt, ax)
    plt.savefig('communication_'+str(num_queries)+'.pdf')

    plt.clf()
    ind = 0
    for scheme in args.name:
        plt.scatter(data[scheme]['item_sz'], data[scheme]['tput'], label="_"+scheme, marker='.')#, c=colors[ind])
        plt.plot(data[scheme]['item_sz'], data[scheme]['tput'], label=scheme)#, c=colors[ind])
        plt.fill_between(data[scheme]['item_sz'], 
                         [x-y for (x,y) in zip(data[scheme]['tput'],data[scheme]['tput_dev'])],
                         [x+y for (x,y) in zip(data[scheme]['tput'],data[scheme]['tput_dev'])],
                         alpha=0.5)
        plt.text(800, 1.08*data[scheme]['total_comm'][11], scheme, size='13')
        ind += 1

    for scheme in tput:
        t = float(tput[scheme])
        plt.axhline(y=t, linestyle='-', c="gray")
        plt.text(1.5, t*1.05, scheme, size='13')
        ind += 1

    plt.xlim(1, 4096)
    #plt.legend()
    plt.xlabel("Entry size (in bits)")
    plt.ylabel("Server throughput (MB/s)")
    plt.xscale('log')
    plt.yscale('log')
    plt.tight_layout()
    custom_style.remove_chart_junk(plt, ax)
    plt.savefig('tput_'+str(num_queries)+'.pdf')

def throughput_with_batching(args):
    data = {}

    fig, ax = plt.subplots()
    #plt.gca().xaxis.set_minor_formatter(FuncFormatter(xaxis))
    ax.yaxis.set_minor_formatter(gbs)
    ax.yaxis.set_major_formatter(gbs)
    ax.xaxis.set_minor_formatter(mticker.ScalarFormatter())
    ax.xaxis.set_major_formatter(mticker.ScalarFormatter())

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

        plt.scatter(data[scheme]['batch_sz'], data[scheme]['tput'], label="_"+scheme, marker='.')#c=colors[ind])
        line, = plt.plot(data[scheme]['batch_sz'], data[scheme]['tput'], label=scheme, marker='.')#,c=colors[ind])
        plt.fill_between(data[scheme]['batch_sz'], 
                         [x-y for (x,y) in zip(data[scheme]['tput'], data[scheme]['dev'])], 
                         [x+y for (x,y) in zip(data[scheme]['tput'], data[scheme]['dev'])], 
                         label="_"+scheme, alpha=0.5, facecolor=line.get_color())
        ind += 1

    #plt.legend()
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
    plt.annotate("DoublePIR", (data['DoublePIR']['batch_sz'][6], 
        data['DoublePIR']['tput'][6]*1.6),
                 size=16, ha='left', rotation=0)
    ax.xaxis.set_major_formatter(FuncFormatter(lambda x,pos: ("%d" % x)))
    ax.yaxis.set_major_formatter(FuncFormatter(lambda x,pos: ("%d" % (x/1024))))
    custom_style.remove_chart_junk(plt, ax)
    fig.set_size_inches([4, 3])
    plt.tight_layout()
    plt.savefig('throughput_with_batching.pdf')

def old(args):
    
    bws = {}
    rates = {}
    heights = {}
    qs = {}

    with open(args.file) as f:
        reader = csv.reader(f)
        next(reader) # skip first line

        for row in reader:
            print(row)
            n = int(row[0])
            l = int(row[1])
            #m = int(row[2])
            q = int(row[3])
            rate = float(row[4])
            bw = float(row[5])

            if n not in bws:
                bws[n] = []
                rates[n] = []
                heights[n] = []
                qs[n] = []

            bws[n].append(bw)
            rates[n].append(rate)
            heights[n].append(l)
            qs[n].append(q)

    plt.clf()
    for n in bws:
        plt.scatter(bws[n], rates[n], label="n=2^"+str(n))
        for i in range(len(bws[n])):
            plt.annotate("l=2^"+str(heights[n][i]),
                         (bws[n][i], rates[n][i]),
                         xytext=(0, 1),
                         textcoords="offset points",
                         ha='center',
                         va='bottom')
    plt.legend()
    plt.xlabel("Communication, for 1 offline and 1 online phase (KB)")
    plt.ylabel("Throughput (MB/s)")
    plt.title(args.name)
    plt.tight_layout()
    custom_style.remove_chart_junk(plt, ax)
    plt.savefig(args.name + ".pdf")


def main(args):
    if args.plot == "batch_tput":
        return throughput_with_batching(args)
    if args.plot == "comm":
        return communication(args)
    if args.plot == "all":
        compare_all(args)
        return plot_tradeoff(args)

if __name__ == "__main__":
    main(args)
