var render = function() {
  var $dateFormat = d3.time.format("%Y-%m-%d");

  var authorStats = function(data) {
    var byAuthor = d3.map()
    data.pull_requests.forEach(function (pr) {
      if (!byAuthor.has(pr.author_ldap)) {
        byAuthor.set(pr.author_ldap, {
          prs: [],
          total_comments_left: 0,
          others_prs_commented_in: 0,
          others_prs_approved: 0,
          count_merged_by_last_updated_date: {},
          full_name: pr.author_fullname});
      }
      byAuthor.get(pr.author_ldap).prs.push(pr);
      console.log(pr);
    });

    data.pull_requests.forEach(function(pr) {
      for (commenter in pr.comments_by_author_ldap) {
        if (byAuthor.get(commenter)) { // if someone we care about
          byAuthor.get(commenter).total_comments_left += pr.comments_by_author_ldap[commenter];
          if (pr.author_ldap != commenter) {
            byAuthor.get(commenter).others_prs_commented_in += 1;
          }
        }
      }
      for (approver in pr.approvals_by_author_ldap) {
        if (byAuthor.get(approver) && pr.author_ldap != approver) {
          byAuthor.get(approver).others_prs_approved += 1;
        }
      }
      if (pr.state == "MERGED") {
        var whenDone = $dateFormat(new Date(pr.updated_datetime * 1000));
        if (!byAuthor.get(pr.author_ldap).count_merged_by_last_updated_date[whenDone]) {
          byAuthor.get(pr.author_ldap).count_merged_by_last_updated_date[whenDone] = 0;
        }
        byAuthor.get(pr.author_ldap).count_merged_by_last_updated_date[whenDone] += 1;
      }
    });

    byAuthor.forEach(function(k,v) {
      v.prs.sort(function(a,b){
        return b.created_datetime - a.created_datetime;
      });
    });
    return byAuthor;
  };

  var renderAuthorHeader = function(author, stats) {
    var personContainer = d3.selectAll(".teamlisting")
        .append("div")
        .attr("class", "person");
    personContainer.append("h4").text(stats.full_name);

    var merged = stats.prs.filter(function(pr) { return pr.state === "MERGED";}).length;

    var ul = personContainer.append("ul");
    ul.append("li").text("" + stats.prs.length + " PRs").append("small").text(" (" + merged + " merged)");
    ul.append("li").text("Commented in " + stats.others_prs_commented_in + " other people's PRs ")
        .append("small").text("(Approved " + stats.others_prs_approved + " PR's, left "
        + stats.total_comments_left + " total comments)");
    return personContainer;
  };

  var renderPrs = function(byAuthor, personContainer, stats) {
    var width = 150,
        height = 75;

    stats.prs.forEach(function(onePr) {
      /* Creates 2 fixed columns of nodes:
       PR       comment authors
       |      /----- A1
       o------------ A2
       |      \----- A3

       The comment authors are sorted by # of comments. */
      var nodes = [];
      var prNode = { title: onePr.title, type: "PR", root: onePr};
      var links = [];
      for (var author in onePr.comments_by_author_ldap) {
        var count = onePr.comments_by_author_ldap[author];
        var is_approval = onePr.approvals_by_author_ldap[author] > 0;
        var authorNode = {
          title: (count > 1 ? "" + count + " comments" : "" + count + " comment") + " by " + author +
          (is_approval ? " (approval)" : ""),
          type: "comments", author: author, is_approval: is_approval, count: count
        };
        nodes.push(authorNode);
        links.push({source: authorNode, target: prNode, value: authorNode.count});
      }
      nodes.sort(function(a,b) {
        return a.count - b.count;
      });
      nodes.unshift(prNode);

      //console.log(nodes);
      //console.log(links);

      var force = d3.layout.force()
          .charge(-100)
          .linkDistance(20)
          .size([width, height]);

      force.nodes(nodes)
          .links(links)
          .start();

      var prAnchor = personContainer
          .append("div")
          .attr("class", "pr-container")
          .append("a")
          .attr("href", onePr.self_url)
      prAnchor
          .append("small").text(onePr.title);
      if (onePr.state == "OPEN") {
        prAnchor.append("span").attr("class", "label label-danger small").text("OPEN");
      }

      var svg = prAnchor
          .append("svg")
          .attr("width", width)
          .attr("height", height);

      var link = svg.selectAll(".link")
          .data(links)
          .enter().append("line")
          .attr("class", "link")
          .style("stroke-width", function(d) { return 1.5 * d.value + "px";});

      var node = svg.selectAll(".node")
          .data(nodes)
          .enter().append("circle")
          .attr("class", "node")
          .attr("r", 5)
          .style("fill", function(n) {
            if (n.type == "PR") {
              return "orange";
            }
            return "blue";
          });

      node.attr("cx", function(d) {
        switch (d.type) {
          case "PR":
            return 10;
          case "comments":
            return 100;
          default:
            throw "error in d.type";
        }
      });

      node.append("title")
          .text(function(d) {
            return d.title; })

      force.on("tick", function() {
        link.attr("x1", function(d) { return d.source.x; })
            .attr("y1", function(d) { return d.source.y; })
            .attr("x2", function(d) { return d.target.x; })
            .attr("y2", function(d) { return d.target.y; });
        node.attr("cx", function(d) { return d.x; })
            .attr("cy", function(d) { return d.y; });
      });
    })
  }

  var renderCalendar = function(personContainer, author, stats) {
    // Inspired by http://bl.ocks.org/mbostock/4063318
    var width = 950, height = 130, cellSize = 17;

    var svg = personContainer.selectAll("svg.calendar")
        .data(d3.range(2015, 2016))
      .enter().append("svg")
        .attr("class", "calendar RdYlGn")
        .attr("width", width)
        .attr("height", height)
      .append("g")
        .attr("transform", "translate(" + ((width - cellSize * 53) / 2) + "," + (height - cellSize * 7 - 1) + ")");

    svg.append("text")
        .attr("transform", "translate(-6," + cellSize * 3.5 + ")rotate(-90)")
        .style("text-anchor", "middle")
        .text(function(d) { return d; });

    var rect = svg.selectAll(".day")
        .data(function(d) { return d3.time.days(new Date(d, 0, 1), new Date(d + 1, 0, 1)); })
        .enter().append("rect")
        .attr("class", "day")
        .attr("width", cellSize)
        .attr("height", cellSize)
        .attr("x", function(d) { return d3.time.weekOfYear(d) * cellSize; })
        .attr("y", function(d) { return d.getDay() * cellSize; })
        .datum($dateFormat);

    rect.append("title")
        .text(function(d) { return d; });

    svg.selectAll(".month")
        .data(function(d) { return d3.time.months(new Date(d, 0, 1), new Date(d + 1, 0, 1)); })
        .enter().append("path")
        .attr("class", "month")
        .attr("d", monthPath);

    var color = d3.scale.quantize()
        .domain([0, 6])// TODO find max # PR's, but this now is probably crazy high enough.
        .range(d3.range(11).map(function(d) { return "q" + d + "-11"; }));

    rect.filter(function(d) { return d in stats.count_merged_by_last_updated_date })
        .attr("class",
            function(d) { return "day " + color(stats.count_merged_by_last_updated_date[d]); })
        .select("title")
        .text(function(d) { return d + ": " +
            stats.count_merged_by_last_updated_date[d] + " merged"; })

    function monthPath(t0) {
      var t1 = new Date(t0.getFullYear(), t0.getMonth() + 1, 0),
          d0 = t0.getDay(), w0 = d3.time.weekOfYear(t0),
          d1 = t1.getDay(), w1 = d3.time.weekOfYear(t1);
      return "M" + (w0 + 1) * cellSize + "," + d0 * cellSize
          + "H" + w0 * cellSize + "V" + 7 * cellSize
          + "H" + w1 * cellSize + "V" + (d1 + 1) * cellSize
          + "H" + (w1 + 1) * cellSize + "V" + 0
          + "H" + (w0 + 1) * cellSize + "Z";
    }

  };

  d3.json('/assets/pr-data-pull.json', function(data) {
    var byAuthor = authorStats(data);
    byAuthor.forEach(function(author, stats) {
      var personContainer = renderAuthorHeader(author, stats);
      renderCalendar(personContainer, author, stats);
      renderPrs(byAuthor, personContainer, stats);
    });
  });
};

render();
