FROM redmine

# The gitmike theme can be retrieved from its project page
# Hide Sidebar plugin allows to hide the sidebar. Especially useful when editing issues
# and Clipboard Image Paste allows to attach an image pasting from the clipboard instead of selecting a file
RUN apt update && apt install -y git \
    build-essential imagemagick libmagickcore-dev libmagickwand-dev ruby-dev \
    && gem install rmagick \
    && git clone https://github.com/makotokw/redmine-theme-gitmike.git public/themes/gitmike \
    && git clone https://gitlab.com/bdemirkir/sidebar_hide.git plugins/sidebar_hide \
    && git clone https://github.com/RubyClickAP/clipboard_image_paste.git plugins/clipboard_image_paste \
#New Add
    && git clone https://github.com/farend/redmine_searchable_selectbox.git plugins/redmine_searchable_selectbox \
    && git clone https://github.com/apsmir/custom_field_sql.git plugins/custom_field_sql \
    && git clone https://github.com/onozaty/redmine-view-customize.git plugins/view_customize \
    && git clone https://github.com/anteo/redmine_custom_workflows.git plugins/redmine_custom_workflows

# The A1 theme had to be downloaded from an email link so the theme files were added to this project directly
COPY a1 public/themes/a1
# The PurpleMine theme
COPY PurpleMine2-master public/themes/PurpleMine2-master
# The abacusmine theme
COPY abacusmine_2.0.6 public/themes/abacusmine_2.0.6
#Redmineup Tags
COPY redmineup_tags plugins/redmineup_tags
#Install Patch (Quick Change Status)
COPY select-status.patch /usr/src/redmine
#patch -p1 < /usr/src/redmine/select-status.patch
#Install Patch (欄位拖拉)
#COPY GCF-4.2.patch redmine
COPY GCF-4.2.patch /usr/src/redmine
#patch -p1 < /usr/src/redmine/GCF-4.2.patch

#Using Bundler with Rails在Build Image時先確認GemFile Dependency, 避免Runtime時在內網環境無法access RubyGems.org
RUN bundle install